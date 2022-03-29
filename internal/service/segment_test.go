/**
@author ChenZhiYin 1981330085@qq.com
@date 2021/11/23
*/

package service

import (
	"bbbid/internal/biz"
	"bbbid/internal/conf"
	"bbbid/internal/data"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/durationpb"
	"os"
	"testing"
	"time"
)

func BenchmarkSegmentUsecase_GetId(b *testing.B) {
	b.StopTimer()
	Name := "bbbid"
	Version := "1.0.0"

	logger := log.With(log.NewStdLogger(os.Stdout),
		"service", Name,
		"version", Version,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)

	bc := conf.Bootstrap{
		Server: nil,
		Data: &conf.Data{
			Database: &conf.Data_Database{
				Driver:    "mysql",
				Dsn:       "test:123456@tcp(192.168.0.250:3306)/bbbid?charset=utf8mb4&parseTime=True&loc=Local",
				MaxConns:  100,
				IdleConns: 10,
				LifeTime: &durationpb.Duration{
					Seconds: 1800,
					Nanos:   0,
				},
				IdleTime: &durationpb.Duration{
					Seconds: 600,
					Nanos:   0,
				},
			},
			Redis: &conf.Data_Redis{
				Addr: "192.168.10.250:6379",
				DialTimeout: &durationpb.Duration{
					Seconds: 1,
					Nanos:   0,
				},
				ReadTimeout: &durationpb.Duration{
					Seconds: 1,
					Nanos:   0,
				},
				WriteTimeout: &durationpb.Duration{
					Seconds: 1,
					Nanos:   0,
				},
			},
		},
	}

	db := data.NewDb(bc.GetData())
	client := data.NewRedis(bc.GetData())
	dataData, _, err := data.NewData(db, client, logger)
	if err != nil {
		return
	}
	segmentRepo := data.NewSegmentRepo(dataData, logger)

	uc := biz.NewSegmentUsecase(segmentRepo, logger)

	//mu := sync.Mutex{}
	//errnum := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
		_, err = uc.GetId(ctx, "test")
		if err != nil {
			b.Error(err)
			//mu.Lock()
			//errnum++
			//mu.Unlock()
		}
	}

	//fmt.Println("errnum:", errnum)
}
