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
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/durationpb"
	"os"
	"sync"
	"testing"
	"time"
)

func TestSegmentService_GetId(t *testing.T) {
	fmt.Println("Test GetId")
	//fmt.Println("Benchmark GetQueueId")
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
				Addr: "192.168.0.250:6379",
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
				Cluster: &conf.Data_Redis_Cluster{Addrs: []string{"192.168.0.250:6379", "192.168.0.250:6380", "192.168.0.251:6379", "192.168.0.251:6380", "192.168.0.252:6379", "192.168.0.252:6380"}},
			},
		},
	}

	//fmt.Println(fmt.Sprintf("conf: %+v", bc))

	db := data.NewDb(bc.GetData())
	client := data.NewRedis(bc.GetData())
	dataData, _, err := data.NewData(db, client, logger)
	if err != nil {
		return
	}
	segmentRepo := data.NewSegmentRepo(dataData, logger)

	uc := biz.NewSegmentUsecase(segmentRepo, logger)

	//var rander = rand.New(rand.NewSource(time.Now().UnixNano()))

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	idmap := make(map[int64]bool)
	zeronum := 0
	idnum := 0
	for i := 0; i < 50000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			//n := rander.Intn(1000)
			//time.Sleep(time.Duration(n)*time.Nanosecond)

			ctx, _ := context.WithTimeout(context.Background(), time.Second)
			id, err := uc.GetId(ctx, "test")

			mu.Lock()
			if id == 0 || err != nil {
				zeronum++
			}
			if id > 0 {
				idnum++
				//logger.Log(log.LevelInfo, "ID", id)
				//fmt.Println("id: ", id)
				if _, ok := idmap[id]; ok {
					fmt.Println("exist id: ", id)
				}
				idmap[id] = true
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	fmt.Println("zeronum: ", zeronum)
	fmt.Println("idnum: ", idnum)
	fmt.Println("idmap: ", len(idmap))
}

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
				Addr: "192.168.0.250:6379",
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
				Cluster: &conf.Data_Redis_Cluster{Addrs: []string{"192.168.0.250:6379", "192.168.0.250:6380", "192.168.0.251:6379", "192.168.0.251:6380", "192.168.0.252:6379", "192.168.0.252:6380"}},
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
	errnum := 0
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*1)
		_, err = uc.GetId(ctx, "test")
		if err != nil {
			b.Error(err)
			//mu.Lock()
			errnum++
			//mu.Unlock()
		}
	}

	//fmt.Println("errnum:", errnum)
}
