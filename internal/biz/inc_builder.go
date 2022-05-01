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

var _ IdBuilder = (*IncBuilder)(nil)

type IncBuilder struct {
	repo SegmentRepo
	log  *log.Helper

	running   bool
	currentId int64
	nextMaxId int64

	currentStep Segment

	mu        sync.Mutex
	runningMu *sync.Cond

	rander *rand.Rand
}

func NewIncBuilder(repo SegmentRepo, logger *log.Helper) IdBuilder {
	return &IncBuilder{
		repo:        repo,
		log:         logger,
		currentStep: Segment{},
		mu:          sync.Mutex{},
		runningMu:   sync.NewCond(&sync.Mutex{}),
	}
}

func (b *IncBuilder) Init(ctx context.Context, segment Segment) (err error) {
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

func (b *IncBuilder) Next(ctx context.Context) error {
	maxid, err := b.repo.GetNextMaxid(ctx, &b.currentStep)

	if err != nil {
		b.log.Error("Next err: ", err)
		return err
	}

	b.mu.Lock()
	b.nextMaxId = maxid
	b.mu.Unlock()

	upctx, _ := context.WithTimeout(context.Background(), time.Second*1)
	stepstr, _ := json.Marshal(&Segment{
		Ckey:  b.currentStep.Ckey,
		Type:  b.currentStep.Type,
		Step:  b.currentStep.Step,
		Maxid: maxid,
	})

	//update maxid queue
	_, err = b.repo.PushStep(upctx, UpdateQueueKey, string(stepstr))
	if err != nil {
		b.log.Error("PushStep Err: ", err)
	}

	return nil
}

func (b *IncBuilder) checkUpate() {
	b.runningMu.L.Lock()
	b.mu.Lock()

	if !b.running && float32(b.currentStep.Maxid-b.currentId)/float32(b.currentStep.Step) < 0.5 {
		b.mu.Unlock()

		b.running = true
		b.runningMu.L.Unlock()

		util.Go(func() {
			defer func() {
				b.runningMu.L.Lock()
				b.running = false
				b.runningMu.L.Unlock()
				b.runningMu.Broadcast()
			}()
			qctx, _ := context.WithTimeout(context.Background(), time.Second)
			_ = b.Next(qctx)
		})
	} else {
		b.runningMu.L.Unlock()
		b.mu.Unlock()
	}

}

func (b *IncBuilder) GetId(ctx context.Context) (int64, error) {
	b.mu.Lock()

	if b.currentId == 0 || b.currentId >= b.currentStep.Maxid {
		if b.nextMaxId > 0 {
			b.currentId = b.nextMaxId - b.currentStep.Step
			b.currentStep.Maxid = b.nextMaxId
			b.nextMaxId = 0
		} else {
			b.mu.Unlock()
			b.checkUpate()

			b.runningMu.L.Lock()
			for b.running {
				b.runningMu.Wait()
			}
			b.runningMu.L.Unlock()

			var retry int64
			if ctx.Value("bbbid_retry") != nil {
				retry = ctx.Value("bbbid_retry").(int64)
			}
			retry++
			if retry <= 5 {
				//retry
				ctx = context.WithValue(ctx, "bbbid_retry", retry)
				return b.GetId(ctx)
			}

			return 0, errors.New(http.StatusServiceUnavailable, "503", "Busy")
		}
	}
	b.currentId++
	id := b.currentId
	b.mu.Unlock()
	return id, nil
}

func (b *IncBuilder) AddBiz(ctx context.Context, req *bid.AddReq) (seg Segment, err error) {
	return Segment{}, nil
}
