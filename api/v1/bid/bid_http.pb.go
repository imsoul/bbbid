// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// protoc-gen-go-http v2.0.1

package bid

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

type BidHTTPServer interface {
	AddBiz(context.Context, *AddReq) (*AddRes, error)
	GetId(context.Context, *IdReq) (*IdRes, error)
}

func RegisterBidHTTPServer(s *http.Server, srv BidHTTPServer) {
	r := s.Route("/")
	r.GET("/v1/getId/{ckey}", _Bid_GetId0_HTTP_Handler(srv))
	r.POST("/v1/getId", _Bid_GetId1_HTTP_Handler(srv))
	r.GET("/v1/addBiz", _Bid_AddBiz0_HTTP_Handler(srv))
	r.POST("/v1/addBiz", _Bid_AddBiz1_HTTP_Handler(srv))
}

func _Bid_GetId0_HTTP_Handler(srv BidHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in IdReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/bid.v1.Bid/GetId")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetId(ctx, req.(*IdReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*IdRes)
		return ctx.Result(200, reply)
	}
}

func _Bid_GetId1_HTTP_Handler(srv BidHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in IdReq
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/bid.v1.Bid/GetId")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetId(ctx, req.(*IdReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*IdRes)
		return ctx.Result(200, reply)
	}
}

func _Bid_AddBiz0_HTTP_Handler(srv BidHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in AddReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/bid.v1.Bid/AddBiz")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.AddBiz(ctx, req.(*AddReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*AddRes)
		return ctx.Result(200, reply)
	}
}

func _Bid_AddBiz1_HTTP_Handler(srv BidHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in AddReq
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, "/bid.v1.Bid/AddBiz")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.AddBiz(ctx, req.(*AddReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*AddRes)
		return ctx.Result(200, reply)
	}
}

type BidHTTPClient interface {
	AddBiz(ctx context.Context, req *AddReq, opts ...http.CallOption) (rsp *AddRes, err error)
	GetId(ctx context.Context, req *IdReq, opts ...http.CallOption) (rsp *IdRes, err error)
}

type BidHTTPClientImpl struct {
	cc *http.Client
}

func NewBidHTTPClient(client *http.Client) BidHTTPClient {
	return &BidHTTPClientImpl{client}
}

func (c *BidHTTPClientImpl) AddBiz(ctx context.Context, in *AddReq, opts ...http.CallOption) (*AddRes, error) {
	var out AddRes
	pattern := "/v1/addBiz"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation("/bid.v1.Bid/AddBiz"))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}

func (c *BidHTTPClientImpl) GetId(ctx context.Context, in *IdReq, opts ...http.CallOption) (*IdRes, error) {
	var out IdRes
	pattern := "/v1/getId"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation("/bid.v1.Bid/GetId"))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
