package transport

import (
	"context"
)

type RpcServer struct {
	handler Handler
}

func NewRpcServer(h Handler) *RpcServer {
	return &RpcServer{
		handler: h,
	}
}

// 处理来自其它节点来得连接事件
func (s *RpcServer) PushConnect(ctx context.Context, req *Connect) (*Response, error) {
	if s.handler != nil {
		s.handler.OnConnect(req.AgentId, req.ClientId)
	}
	return &Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// 处理来自其它节点来得断开连接事件
func (s *RpcServer) PushDisconnect(ctx context.Context, req *Disconnect) (*Response, error) {
	if s.handler != nil {
		s.handler.OnDisConnect(req.AgentId, req.ClientId)
	}
	return &Response{
		Code: 0,
		Msg:  "success",
	}, nil
}

// 处理来自其它节点来得发布事件
func (s *RpcServer) PushPublish(ctx context.Context, req *Publish) (*Response, error) {
	if s.handler != nil {
		s.handler.OnPublish(req.AgentId, req.Topic, req.Payload, byte(req.Qos), req.Retain)
	}
	return &Response{
		Code: 0,
		Msg:  "success",
	}, nil
}
