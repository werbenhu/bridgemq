package transport

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/werbenhu/bridgemq/discovery"
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

func (g *RpcTransport) Join(node *discovery.Agent) {
	if _, ok := g.clients.Load(node.Id); !ok {
		addr := node.Addr + ":" + node.TransmitPort
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

func (g *RpcTransport) Leave(node *discovery.Agent) {
	if c, ok := g.clients.Load(node.Id); ok {
		addr := node.Addr + ":" + node.TransmitPort
		log.Printf("[INFO] agent: %s has left, addr:%s \n", node.Id, addr)
		client := c.(*RpcClient)
		g.clients.Delete(node.Id)
		client.Close()
	}
}

func (g *RpcTransport) Update(node *discovery.Agent) {
	if _, ok := g.clients.Load(node.Id); !ok {
		addr := node.Addr + ":" + node.TransmitPort
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

func (g *RpcTransport) PushConnect(node *discovery.Agent, clientId string) {
	g.clients.Range(func(key any, val any) bool {
		if key.(string) == node.Id {
			return true
		}
		client := val.(*RpcClient)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// log.Printf("[DEBUG] bridge push connect to agent:%s \n", key.(string))
		if _, err := client.PushConnect(ctx, &Connect{
			AgentId:  node.Id,
			ClientId: clientId,
		}); err != nil {
			log.Printf("[ERROR] bridge push connect to agent:%s failed, err:%s\n", key.(string), err.Error())
		}
		return true
	})
}

func (g *RpcTransport) PushDisconnect(node *discovery.Agent, clientId string) {
	g.clients.Range(func(key any, val any) bool {
		if key.(string) == node.Id {
			return true
		}
		client := val.(*RpcClient)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := client.PushDisconnect(ctx, &Disconnect{
			AgentId:  node.Id,
			ClientId: clientId,
		}); err != nil {
			log.Printf("[ERROR] bridge push disconnect to agent:%s failed, err:%s\n", key.(string), err.Error())
		}
		return true
	})
}

func (g *RpcTransport) PushPublish(node *discovery.Agent, topic string, payload []byte, qos byte, retain bool) {
	g.clients.Range(func(key any, val any) bool {
		if key.(string) == node.Id {
			return true
		}
		client := val.(*RpcClient)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := client.PushPublish(ctx, &Publish{
			AgentId: node.Id,
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
