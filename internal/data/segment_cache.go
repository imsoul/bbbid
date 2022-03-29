/**
@author 陈志银 1981330085@qq.com
@date 2022/2/24
*/

package data

import (
	"bbbid/internal/biz"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

func GetBidKey(ckey string) string {
	return fmt.Sprintf("bbbid:%s", ckey)
}

func (r *segmentRepo) BidInc(ctx context.Context, ckey string) (int64, error) {
	return r.data.rdb.Incr(ctx, GetBidKey(ckey)).Result()
}

func (r *segmentRepo) PopStep(ctx context.Context, key string, timeout time.Duration) ([]string, error) {
	return r.data.rdb.BRPop(ctx, timeout, key).Result()
}

func (r *segmentRepo) PushStep(ctx context.Context, key string, value string) (int64, error) {
	return r.data.rdb.LPush(ctx, key, value).Result()
}

func (r *segmentRepo) GetMaxid(ctx context.Context, ckey string) (int64, error) {
	key := GetBidKey(ckey)
	res, err := r.data.rdb.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	maxid, _ := strconv.Atoi(res)
	return int64(maxid), nil
}

func (r *segmentRepo) GetNextMaxid(ctx context.Context, step *biz.Segment) (int64, error) {
	key := GetBidKey(step.Ckey)

	res, err := r.data.rdb.Get(ctx, key).Result()

	if err != nil || res == "" {
		if err != redis.Nil {
			r.log.Error("redis err: ", err)
			return 0, err
		}
		nextmaxid := step.Maxid + step.Step
		_, err = r.data.rdb.Set(ctx, key, nextmaxid, 0).Result()
		if err != nil {
			return 0, err
		}

		return nextmaxid, err
	} else {
		var nextmaxid int64
		// inc step
		nextmaxid, err = r.data.rdb.IncrBy(ctx, key, step.Step).Result()
		if err != nil {
			return 0, err
		}

		return nextmaxid, err
	}
}
