package biz

import (
	"bbbid/api/v1/bid"
	"bbbid/pkg/util"
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var _ IdBuilder = (*RandBuilder)(nil)

type RandBuilder struct {
	repo SegmentRepo
	log  *log.Helper

	running bool

	currentQueue chan int64
	currentStep  Segment

	mu sync.RWMutex

	rander *rand.Rand
}

func NewRandBuilder(repo SegmentRepo, logger *log.Helper) IdBuilder {
	return &RandBuilder{
		repo:        repo,
		log:         logger,
		currentStep: Segment{},
		mu:          sync.RWMutex{},
	}
}

func (b *RandBuilder) Init(ctx context.Context, segment Segment) (err error) {
	b.currentQueue = make(chan int64, segment.Step+segment.Step/2+1)
	b.currentStep = segment

	b.mu.Lock()
	b.running = true
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		b.running = false
		b.mu.Unlock()
	}()

	return b.Next(ctx)
}

func (b *RandBuilder) Next(ctx context.Context) error {
	//TODO: 根据错误率动态调整step
	maxid, err := b.repo.GetNextMaxid(ctx, &b.currentStep)

	if err != nil {
		b.log.Error("Next err: ", err)
		return err
	}

	b.mu.Lock()
	b.currentStep.Maxid = maxid
	b.mu.Unlock()

	upctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	stepstr, _ := json.Marshal(b.currentStep)

	//update maxid queue
	_, err = b.repo.PushStep(upctx, UpdateQueueKey, string(stepstr))
	if err != nil {
		b.log.Error("PushStep Err: ", err)
	}

	nextId := maxid - b.currentStep.Step
	idlist := make([]int64, 0)

	for i := nextId; i < maxid; i++ {
		idlist = append(idlist, i)
	}

	//rand
	rand.Shuffle(len(idlist), func(i, j int) { idlist[i], idlist[j] = idlist[j], idlist[i] })

	for _, id := range idlist {
		b.currentQueue <- id
	}

	return nil
}

func (b *RandBuilder) GetId(ctx context.Context) (int64, error) {
	//fmt.Println("Rand GetId")

	b.mu.RLock()
	isRunning := b.running
	step := b.currentStep.Step
	b.mu.RUnlock()

	if !isRunning && float32(len(b.currentQueue))/float32(step) < 0.5 {
		b.mu.Lock()
		b.running = true
		b.mu.Unlock()
		util.Go(func() {
			defer func() {
				b.mu.Lock()
				b.running = false
				b.mu.Unlock()
			}()

			qctx, _ := context.WithTimeout(context.Background(), time.Second)
			_ = b.Next(qctx)
		})
	}

	for {
		select {
		case <-ctx.Done():
			return 0, nil
		case <-time.After(time.Millisecond * 500):
			return 0, errors.New(http.StatusServiceUnavailable, "503", "Busy")
		case id, ok := <-b.currentQueue:
			if !ok {
				return 0, errors.New(http.StatusNotFound, "404", "Busy")
			}

			return id, nil
		}
	}
}

func (b *RandBuilder) AddBiz(ctx context.Context, req *bid.AddReq) (seg Segment, err error) {
	return Segment{}, nil
}
