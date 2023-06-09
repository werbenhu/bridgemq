package transport

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/werbenhu/bridgemq/agent"
	"google.golang.org/grpc"
)

const (
	AgentIdKey = "agent-id"
)

type Opt struct {
	Port string
}

type RpcTransport struct {
	server  *grpc.Server
	opts    *Opt
	handler Handler
	clients sync.Map
}

func NewRpcTransport(opts *Opt) *RpcTransport {
	return &RpcTransport{
		opts: opts,
	}
}

func (g *RpcTransport) SetHandler(h Handler) {
	g.handler = h
}

// Join() is called When a new agent is discovered by the discovery
func (g *RpcTransport) Join(node *agent.Agent) {
	if _, ok := g.clients.Load(node.Id); !ok {
		addr := node.Addr + ":" + node.PipePort
		log.Printf("[INFO] agent: %s has joined, addr:%s \n", node.Id, addr)
		conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithUserAgent(node.Id))
		if err != nil {
			log.Printf("[ERROR] agent join failed. grpc dial addr:%s err:%s\n", addr, err.Error())
			return
		}

		client := NewTransportClient(conn)
		g.clients.Store(node.Id, &RpcClient{
			conn: conn,
			pipe: client,
		})
	}
}

// Leave() is called When a agent is left
func (g *RpcTransport) Leave(node *agent.Agent) {
	if c, ok := g.clients.Load(node.Id); ok {
		addr := node.Addr + ":" + node.PipePort
		log.Printf("[INFO] agent: %s has left, addr:%s \n", node.Id, addr)
		client := c.(*RpcClient)
		g.clients.Delete(node.Id)
		client.Close()
	}
}

// Join() is called When a agent updated
func (g *RpcTransport) Update(node *agent.Agent) {
	if _, ok := g.clients.Load(node.Id); !ok {
		addr := node.Addr + ":" + node.PipePort
		log.Printf("[INFO] agent: %s was updated, addr: %s \n", node.Id, addr)
		conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithUserAgent(node.Id))
		if err != nil {
			log.Printf("[ERROR] agent update failed. grpc dial addr:%s err:%s\n", addr, err.Error())
			return
		}

		client := NewTransportClient(conn)
		g.clients.Store(node.Id, &RpcClient{
			conn: conn,
			pipe: client,
		})
	}
}

// PushConnect transmit a connect package to the remote agent via grpc
// clientId is the client id of the client that connected
func (g *RpcTransport) PushConnect(local *agent.Agent, clientId string) {
	g.clients.Range(func(key any, val any) bool {
		if local.IsSelf(key.(string)) {
			return true
		}
		client := val.(*RpcClient)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := client.PushConnect(ctx, &Connect{
			AgentId:  local.Id,
			ClientId: clientId,
		}); err != nil {
			log.Printf("[ERROR] bridge push connect to agent:%s failed, err:%s\n", key.(string), err.Error())
		}
		return true
	})
}

// PushDisconnect transmit a connect package to the remote agent via grpc
// clientId is the client id of the client that connected
func (g *RpcTransport) PushDisconnect(local *agent.Agent, clientId string) {
	g.clients.Range(func(key any, val any) bool {
		if local.IsSelf(key.(string)) {
			return true
		}
		client := val.(*RpcClient)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := client.PushDisconnect(ctx, &Disconnect{
			AgentId:  local.Id,
			ClientId: clientId,
		}); err != nil {
			log.Printf("[ERROR] bridge push disconnect to agent:%s failed, err:%s\n", key.(string), err.Error())
		}
		return true
	})
}

// PushPublish transmit a publish package to the remote agent via grpc
func (g *RpcTransport) PushPublish(local *agent.Agent, topic string, payload []byte, qos byte, retain bool) {
	g.clients.Range(func(key any, val any) bool {
		if local.IsSelf(key.(string)) {
			return true
		}
		client := val.(*RpcClient)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := client.PushPublish(ctx, &Publish{
			AgentId: local.Id,
			Topic:   topic,
			Payload: payload,
			Qos:     int32(qos),
			Retain:  retain,
		}); err != nil {
			log.Printf("[ERROR] bridge push publish to agent:%s failed, err:%s\n", key.(string), err.Error())
		}
		return true
	})
}

func (g *RpcTransport) Start() error {
	var err error

	listener, err := net.Listen("tcp", ":"+g.opts.Port)
	if err != nil {
		log.Fatalf("[ERROR] rpc transport listen to port:%s failed, err:%s", g.opts.Port, err.Error())
		return err
	}

	g.server = grpc.NewServer()
	RegisterTransportServer(g.server, NewRpcServer(g.handler))
	if err = g.server.Serve(listener); err != nil {
		log.Fatalf("[ERROR] rpc transport serve to port:%s failed, err:%s", g.opts.Port, err.Error())
	}
	return err
}

func (g *RpcTransport) Stop() {
	g.clients.Range(func(k any, v any) bool {
		v.(*RpcClient).Close()
		return true
	})
	g.server.Stop()
}
