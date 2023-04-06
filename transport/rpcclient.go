package transport

import (
	"context"

	"google.golang.org/grpc"
)

type RpcClient struct {
	conn *grpc.ClientConn
	pipe TransportClient
}

func (c *RpcClient) Close() {
	c.conn.Close()
}

// PushConnect send a connect package to the remote agent via grpc
func (c *RpcClient) PushConnect(ctx context.Context, in *Connect, opts ...grpc.CallOption) (*Response, error) {
	return c.pipe.PushConnect(ctx, in, opts...)
}

// PushDisconnect send a disconnect package to the remote agent via grpc
func (c *RpcClient) PushDisconnect(ctx context.Context, in *Disconnect, opts ...grpc.CallOption) (*Response, error) {
	return c.pipe.PushDisconnect(ctx, in, opts...)
}

// PushPublish send a publish package to the remote agent via grpc
func (c *RpcClient) PushPublish(ctx context.Context, in *Publish, opts ...grpc.CallOption) (*Response, error) {
	return c.pipe.PushPublish(ctx, in, opts...)
}
