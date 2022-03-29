/**
@author 陈志银 1981330085@qq.com
@date 2021/11/23
*/

package data

import (
	"bbbid/internal/biz"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type segmentRepo struct {
	data *Data
	log  *log.Helper
}

// NewSegmentRepo .
func NewSegmentRepo(data *Data, logger log.Logger) biz.SegmentRepo {
	return &segmentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *segmentRepo) CreateSegment(ctx context.Context, seg *biz.Segment) (err error) {
	err = r.data.db.WithContext(ctx).Model(&biz.Segment{}).Create(&seg).Error
	return
}

func (r *segmentRepo) UpdateMaxid(ctx context.Context, ckey string, maxid int64) (err error) {
	mutexname := fmt.Sprintf("bbbid:lock:%s", ckey)
	mutex := r.data.rsync.NewMutex(mutexname)

	// get lock
	if err = mutex.LockContext(ctx); err != nil {
		return err
	}

	defer func() {
		if ok, err1 := mutex.UnlockContext(ctx); !ok || err1 != nil {
			r.log.Error("unlock failed")
		}
	}()

	return r.data.db.WithContext(ctx).Model(&biz.Segment{}).Where("ckey = ? and maxid < ?", ckey, maxid).UpdateColumn("maxid", maxid).Error
}

func (r *segmentRepo) UpdateSegment(ctx context.Context, ckey string) (segment biz.Segment, err error) {
	db := r.data.db.WithContext(ctx)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Error; err != nil {
		return
	}

	err = tx.Set("gorm:query_option", "FOR UPDATE").Where("ckey = ?", ckey).First(&segment).Error

	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Model(&biz.Segment{}).Where("ckey = ?", ckey).UpdateColumn("maxid", gorm.Expr("maxid + step")).Error
	if err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Commit().Error; err != nil {
		return
	}

	segment.Maxid += segment.Step

	return
}

func (r *segmentRepo) GetSegment(ctx context.Context, ckey string) (segment biz.Segment, err error) {
	err = r.data.db.WithContext(ctx).Model(&biz.Segment{}).Where("ckey = ?", ckey).First(&segment).Error
	return
}

func (r *segmentRepo) GetSegmentList(ctx context.Context) (list []biz.Segment, err error) {
	err = r.data.db.WithContext(ctx).Model(&biz.Segment{}).Find(&list).Error
	return
}
