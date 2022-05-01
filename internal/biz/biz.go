package biz

import (
	"bbbid/api/v1/bid"
	"context"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewSegmentUsecase)

type IdBuilder interface {
	Init(ctx context.Context, segment Segment) (err error)
	Next(ctx context.Context) error
	GetId(ctx context.Context) (int64, error)
	AddBiz(ctx context.Context, req *bid.AddReq) (seg Segment, err error)
}
