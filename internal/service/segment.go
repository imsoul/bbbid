/**
@author ChenZhiYin 1981330085@qq.com
@date 2021/11/23
*/

package service

import (
	pb "bbbid/api/v1/bid"
	"bbbid/internal/biz"
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type SegmentService struct {
	pb.UnimplementedBidServer

	uc  *biz.SegmentUsecase
	log *log.Helper
}

func NewSegmentService(uc *biz.SegmentUsecase, logger log.Logger) *SegmentService {
	return &SegmentService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// AddBiz 增加业务key
func (s *SegmentService) AddBiz(ctx context.Context, req *pb.AddReq) (res *pb.AddRes, err error) {
	_, err = s.uc.AddBiz(ctx, req)
	if err != nil {
		s.log.Error("AddBiz", err)
		return nil, err
	}

	return &pb.AddRes{}, nil
}

// GetId 获取ID
func (s *SegmentService) GetId(ctx context.Context, req *pb.IdReq) (res *pb.IdRes, err error) {
	id, err := s.uc.GetId(ctx, req.GetCkey())
	if err != nil || id == 0 {
		//s.log.Errorf("id: %d, err: %+v", id, err)
		return nil, err
	}

	return &pb.IdRes{
		Id: id,
	}, nil
}
