/**
@author ChenZhiYin 1981330085@qq.com
@date 2021/11/23
*/

package data

import (
	"bbbid/internal/conf"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDb, NewRedis, NewSegmentRepo)

// Data .
type Data struct {
	db    *gorm.DB
	rdb   *redis.Client
	rsync *redsync.Redsync
}

// NewData .
func NewData(db *gorm.DB, rdb *redis.Client, logger log.Logger) (data *Data, cleanup func(), err error) {
	cleanup = func() {
		mylog := log.NewHelper(logger)
		mylog.Info("cleanup bid data resources")

		if db != nil {
			dbc, _ := db.DB()
			if dbc != nil {
				_ = dbc.Close()
			}
		}
		if rdb != nil {
			_ = rdb.Close()
		}
	}

	pool := goredis.NewPool(rdb) // or, pool := redigo.NewPool(...)

	return &Data{
		db:    db,
		rdb:   rdb,
		rsync: redsync.New(pool),
	}, cleanup, err
}

// NewDb init db
func NewDb(conf *conf.Data) *gorm.DB {
	dbClient, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       conf.Database.Dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	}), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	db, err := dbClient.DB()
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(int(conf.Database.GetIdleConns()))
	db.SetMaxOpenConns(int(conf.Database.GetMaxConns()))
	db.SetConnMaxLifetime(conf.Database.GetLifeTime().AsDuration())
	db.SetConnMaxIdleTime(conf.Database.GetIdleTime().AsDuration())

	return dbClient
}

// NewRedis init redis
func NewRedis(conf *conf.Data) *redis.Client {
	rdbClient := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           int(conf.Redis.Db),
		DialTimeout:  conf.Redis.DialTimeout.AsDuration(),
		WriteTimeout: conf.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  conf.Redis.ReadTimeout.AsDuration(),
	})
	rdbClient.AddHook(redisotel.TracingHook{})

	_, err := rdbClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	return rdbClient
}
