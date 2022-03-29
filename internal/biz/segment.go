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
	Running bool   `gorm:"-"`
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

type SegmentUsecase struct {
	repo SegmentRepo
	log  *log.Helper

	currentQueue map[string]chan int64
	currentStep  map[string]*Segment

	queueMu sync.RWMutex

	rander *rand.Rand
}

func NewSegmentUsecase(repo SegmentRepo, logger log.Logger) *SegmentUsecase {
	useCase := &SegmentUsecase{
		repo:         repo,
		log:          log.NewHelper(logger),
		queueMu:      sync.RWMutex{},
		currentQueue: make(map[string]chan int64),
		currentStep:  make(map[string]*Segment),
		rander:       rand.New(rand.NewSource(time.Now().UnixNano())),
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

func (uc *SegmentUsecase) Init(ctx context.Context) (err error) {
	segmentList, err := uc.repo.GetSegmentList(ctx)
	if err != nil {
		return
	}
	if len(segmentList) > 0 {
		for _, segment := range segmentList {
			err = uc.InitQueue(ctx, segment)
			if err != nil {
				continue
			}
		}
	}
	return nil
}

func (uc *SegmentUsecase) InitQueue(ctx context.Context, segment Segment) error {
	uc.currentQueue[segment.Ckey] = make(chan int64, segment.Step+segment.Step/2+1)
	uc.currentStep[segment.Ckey] = &segment

	uc.queueMu.Lock()
	uc.currentStep[segment.Ckey].Running = true
	uc.queueMu.Unlock()

	defer func() {
		uc.queueMu.Lock()
		uc.currentStep[segment.Ckey].Running = false
		uc.queueMu.Unlock()
	}()

	return uc.NextQueue(ctx, segment.Ckey)
}

func (uc *SegmentUsecase) NextQueue(ctx context.Context, ckey string) error {
	//TODO: 根据错误率动态调整step
	maxid, err := uc.repo.GetNextMaxid(ctx, uc.currentStep[ckey])

	if err != nil {
		uc.log.Error("Next err: ", err)
		return err
	}

	uc.currentStep[ckey].Maxid = maxid

	upctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	stepstr, _ := json.Marshal(uc.currentStep[ckey])

	//update maxid queue
	_, err = uc.repo.PushStep(upctx, UpdateQueueKey, string(stepstr))
	if err != nil {
		uc.log.Error("PushStep Err: ", err)
	}

	nextId := maxid - uc.currentStep[ckey].Step

	if uc.currentStep[ckey].Type == int8(bid.BidType_BIDTYPE_INCREMENT) {
		for i := nextId; i < maxid; i++ {
			uc.currentQueue[ckey] <- i
		}
	} else if uc.currentStep[ckey].Type == int8(bid.BidType_BIDTYPE_RAND) {
		idlist := make([]int64, 0)

		for i := nextId; i < maxid; i++ {
			idlist = append(idlist, i)
		}

		//rand
		rand.Shuffle(len(idlist), func(i, j int) { idlist[i], idlist[j] = idlist[j], idlist[i] })

		for _, id := range idlist {
			uc.currentQueue[ckey] <- id
		}
	} else {
		return errors.New(http.StatusNotFound, "404", "Type Error")
	}

	return nil
}

func (uc *SegmentUsecase) GetId(ctx context.Context, ckey string) (int64, error) {
	if _, ok := uc.currentQueue[ckey]; !ok {
		return 0, nil
	}

	uc.queueMu.RLock()
	isRunning := uc.currentStep[ckey].Running
	step := uc.currentStep[ckey].Step
	uc.queueMu.RUnlock()

	if !isRunning && float32(len(uc.currentQueue[ckey]))/float32(step) < 0.5 {
		uc.queueMu.Lock()
		uc.currentStep[ckey].Running = true
		uc.queueMu.Unlock()
		util.Go(func() {
			defer func() {
				uc.queueMu.Lock()
				uc.currentStep[ckey].Running = false
				uc.queueMu.Unlock()
			}()

			qctx, _ := context.WithTimeout(context.Background(), time.Second)
			_ = uc.NextQueue(qctx, ckey)
		})
	}

	for {
		select {
		case <-ctx.Done():
			return 0, nil
		case <-time.After(time.Millisecond * 500):
			return 0, errors.New(http.StatusServiceUnavailable, "503", "Busy")
		case id, ok := <-uc.currentQueue[ckey]:
			if !ok {
				return 0, errors.New(http.StatusNotFound, "404", "Busy")
			}

			return id, nil
		}
	}
}

// AddBiz 新增业务
func (uc *SegmentUsecase) AddBiz(ctx context.Context, req *bid.AddReq) (seg Segment, err error) {
	seg, err = uc.repo.GetSegment(ctx, req.GetCkey())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	if seg.Ckey != "" {
		err = errors.New(http.StatusBadRequest, "400", "Key Exist")
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

	err = uc.InitQueue(ctx, seg)
	return
}
