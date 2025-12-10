package gateway

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type pt[T any] interface {
	proto.Message
	*T
}

type handler[Req pt[T], Resp pt[K], T, K any] func(context.Context, Req) (Resp, error)

type Adapter interface {
	GetMethod() string
	Handle(ctx context.Context, connect grpc.ClientConnInterface, body []byte) (*Response, error)
}

func NewAdapter[Req pt[T], Resp pt[K], T, K any](method string, _ handler[Req, Resp, T, K]) Adapter {
	return &protoAdapter[Req, Resp, T, K]{method: method}
}

type protoAdapter[Req pt[T], Resp pt[K], T, K any] struct {
	method string
}

func (a *protoAdapter[Req, Resp, T, K]) GetMethod() string {
	return a.method
}

func (a *protoAdapter[Req, Resp, T, K]) Handle(ctx context.Context, connect grpc.ClientConnInterface, body []byte) (*Response, error) {
	req := Req(new(T))
	err := proto.Unmarshal(body, req)
	if err != nil {
		return nil, err
	}

	var metadata serverMetadata

	resp := Resp(new(K))
	err = connect.Invoke(ctx, a.method, req, resp, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, err
		}
		return &Response{Body: nil, Status: st, Metadata: metadata.HeaderMD}, nil
	}

	res, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return &Response{Body: res, Status: nil, Metadata: metadata.HeaderMD}, nil
}
