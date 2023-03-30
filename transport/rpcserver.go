package transport

import (
	"context"
)

// RpcServer is a grpc server recive connet, disconnect and publish package from other agent
type RpcServer struct {
	handler Handler
}

func NewRpcServer(h Handler) *RpcServer {
	return &RpcServer{
		handler: h,
	}
}

// PushConnect handle connect package from other agents via grpc
func (s *RpcServer) PushConnect(ctx context.Context, req *Connect) (*Response, error) {
	if s.handler != nil {
		s.handler.OnConnect(req.AgentId, req.ClientId)
	}
	return &Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// PushDisconnect handle disconnect package from other agents via grpc
func (s *RpcServer) PushDisconnect(ctx context.Context, req *Disconnect) (*Response, error) {
	if s.handler != nil {
		s.handler.OnDisConnect(req.AgentId, req.ClientId)
	}
	return &Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// PushPublish handle publich package from other agents via grpc
func (s *RpcServer) PushPublish(ctx context.Context, req *Publish) (*Response, error) {
	if s.handler != nil {
		s.handler.OnPublish(req.AgentId, req.Topic, req.Payload, byte(req.Qos), req.Retain)
	}
	return &Response{
		Code: 0,
		Msg:  "success",
	}, nil
}
