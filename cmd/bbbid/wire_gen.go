// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"bbbid/internal/biz"
	"bbbid/internal/conf"
	"bbbid/internal/data"
	"bbbid/internal/server"
	"bbbid/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// Injectors from wire.go:

// initApp init kratos application.
func initApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	db := data.NewDb(confData)
	client := data.NewRedis(confData)
	dataData, cleanup, err := data.NewData(db, client, logger)
	if err != nil {
		return nil, nil, err
	}
	segmentRepo := data.NewSegmentRepo(dataData, logger)
	segmentUsecase := biz.NewSegmentUsecase(segmentRepo, logger)
	segmentService := service.NewSegmentService(segmentUsecase, logger)
	httpServer := server.NewHTTPServer(confServer, segmentService, logger)
	grpcServer := server.NewGRPCServer(confServer, segmentService, logger)
	app := newApp(logger, httpServer, grpcServer)
	return app, func() {
		cleanup()
	}, nil
}