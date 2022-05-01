/**
@author ChenZhiYin 1981330085@qq.com
@date 2021/11/23
*/

package biz

import (
	"bbbid/api/v1/bid"
	"bbbid/pkg/util"
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Segment struct {
	Ckey    string `gorm:"primaryKey;autoIncrement:true" json:"ckey,omitempty"`
	Type    int8   `json:"type,omitempty"`
	Step    int64  `json:"step,omitempty"`
	Maxid   int64  `json:"maxid,omitempty"`
	Intro   string `json:"intro,omitempty"`
	Uptime  int64  `json:"uptime,omitempty"`
	Addtime int64  `json:"addtime,omitempty"`
}

func (Segment) TableName() string {
	return "bbbid.segment"
}

const UpdateQueueKey = "bbbid:upsegment"

type SegmentRepo interface {
	CreateSegment(ctx context.Context, seg *Segment) error
	UpdateSegment(ctx context.Context, ckey string) (Segment, error)
	UpdateMaxid(ctx context.Context, ckey string, maxid int64) (err error)
	GetMaxid(ctx context.Context, ckey string) (int64, error)
	GetNextMaxid(ctx context.Context, step *Segment) (int64, error)
	GetSegment(ctx context.Context, ckey string) (Segment, error)
	GetSegmentList(ctx context.Context) ([]Segment, error)
	PopStep(ctx context.Context, key string, timeout time.Duration) ([]string, error)
	PushStep(ctx context.Context, key string, value string) (int64, error)
}

type HandlerFunc func(repo SegmentRepo, logger *log.Helper) IdBuilder

var BuilderMap = map[bid.BidType]HandlerFunc{
	bid.BidType_BIDTYPE_INCREMENT: NewIncBuilder,
	bid.BidType_BIDTYPE_RAND:      NewRandBuilder,
}

type SegmentUsecase struct {
	repo SegmentRepo
	log  *log.Helper

	builderMap map[string]IdBuilder

	mu     sync.RWMutex
	rander *rand.Rand
}

func NewSegmentUsecase(repo SegmentRepo, logger log.Logger) *SegmentUsecase {
	useCase := &SegmentUsecase{
		repo:       repo,
		log:        log.NewHelper(logger),
		mu:         sync.RWMutex{},
		builderMap: make(map[string]IdBuilder),
		rander:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	useCase.UpdateSegmentQueue()

	err := useCase.Init(context.Background())
	if err != nil {
		return nil
	}

	return useCase
}

// UpdateSegmentQueue update maxid to db
func (uc *SegmentUsecase) UpdateSegmentQueue() {
	util.Go(func() {
		var uplist []Segment
		var uptime = time.Now().Unix()

		for {
			//pop update queue
			res, err := uc.repo.PopStep(context.Background(), UpdateQueueKey, 2*time.Second)

			if err != nil {
				if err != redis.Nil {
					uc.log.Error("PopStep err: ", err)
				}
			} else if res[1] != "" {
				//fmt.Println("PopStep: ", res[1])

				step := Segment{}
				err = json.Unmarshal([]byte(res[1]), &step)
				if err != nil {
					continue
				}

				uplist = append(uplist, step)
			}

			timenow := time.Now().Unix()
			uplen := len(uplist)
			if uplen > 10 || (uplen > 0 && timenow-uptime >= 1) {
				uptime = timenow

				upmap := make(map[string]Segment)
				for _, step := range uplist {
					if _, ok := upmap[step.Ckey]; !ok {
						upmap[step.Ckey] = step
					} else {
						if step.Maxid > upmap[step.Ckey].Maxid {
							upmap[step.Ckey] = step
						}
					}
				}

				for _, step := range upmap {
					ctx, _ := context.WithTimeout(context.Background(), time.Second*1)

					err = uc.repo.UpdateMaxid(ctx, step.Ckey, step.Maxid)
					if err != nil {
						stepstr, _ := json.Marshal(step)
						_, err = uc.repo.PushStep(ctx, step.Ckey, string(stepstr))
						if err != nil {
							uc.log.Error("Re PushStep Err: ", err)
						}
					}
				}
				uplist = make([]Segment, 0)
			}
		}
	})
}

func (uc *SegmentUsecase) GetBuilder(bidType bid.BidType) (IdBuilder, error) {
	if _, ok := BuilderMap[bidType]; !ok {
		return nil, errors.New(http.StatusNotFound, "404", "builder not found")
	}
	newFun := BuilderMap[bidType]
	return newFun(uc.repo, uc.log), nil
}

func (uc *SegmentUsecase) Init(ctx context.Context) (err error) {
	segmentList, err := uc.repo.GetSegmentList(ctx)
	if err != nil {
		return
	}
	if len(segmentList) > 0 {
		for _, segment := range segmentList {
			err = uc.InitBuilder(segment)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (uc *SegmentUsecase) InitBuilder(segment Segment) (err error) {
	//fmt.Println("========InitBuilder=======")
	uc.mu.Lock()
	defer uc.mu.Unlock()

	if _, ok := uc.builderMap[segment.Ckey]; ok {
		err = errors.New(http.StatusBadRequest, "400", "Key Exist")
		return
	}

	uc.builderMap[segment.Ckey], err = uc.GetBuilder(bid.BidType(segment.Type))
	if err != nil {
		return err
	}

	//fmt.Println("uc.builderMap: ", uc.builderMap)

	err = uc.builderMap[segment.Ckey].Init(context.Background(), segment)
	if err != nil {
		return
	}

	return nil
}

func (uc *SegmentUsecase) GetId(ctx context.Context, ckey string) (int64, error) {
	uc.mu.RLock()
	builder, ok := uc.builderMap[ckey]
	uc.mu.RUnlock()

	//fmt.Printf("builder: %+v \n", builder)

	if !ok {
		return 0, nil
	}

	return builder.GetId(ctx)
}

// AddBiz 新增业务
func (uc *SegmentUsecase) AddBiz(ctx context.Context, req *bid.AddReq) (seg Segment, err error) {
	if _, ok := uc.builderMap[req.GetCkey()]; ok {
		err = errors.New(http.StatusBadRequest, "400", "Key Exist")
		return
	}

	seg, err = uc.repo.GetSegment(ctx, req.GetCkey())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	if seg.Ckey != "" {
		err = uc.InitBuilder(seg)
		return
	}

	timenow := time.Now().Unix()
	seg = Segment{
		Ckey:    req.GetCkey(),
		Type:    int8(req.GetType()),
		Step:    req.GetStep(),
		Maxid:   req.GetMaxid(),
		Intro:   req.GetIntro(),
		Uptime:  timenow,
		Addtime: timenow,
	}

	err = uc.repo.CreateSegment(ctx, &seg)
	if err != nil {
		return
	}

	err = uc.InitBuilder(seg)

	return
}
