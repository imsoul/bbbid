package server

import (
	bidapi "bbbid/api/v1/bid"
	"bbbid/internal/conf"
	"bbbid/internal/service"
	"bbbid/pkg/bmiddle"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer Create a HTTP server.
func NewHTTPServer(c *conf.Server, us *service.SegmentService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			//tracing.Server(),
			//logging.Server(logger), // 开启日志会导致性能下降
			//metrics.Server(),
			validate.Validator(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	opts = append(opts, http.ResponseEncoder(bmiddle.ResponseSuccess), http.ErrorEncoder(bmiddle.ResponseError))

	srv := http.NewServer(opts...)
	bidapi.RegisterBidHTTPServer(srv, us)

	return srv
}
