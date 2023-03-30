package bridge

import (
	"log"

	"github.com/mochi-co/mqtt/v2/packets"
	"github.com/werbenhu/bridgemq/discovery"
	"github.com/werbenhu/bridgemq/transport"
)

type Bridge struct {
	option    *Option
	discovery discovery.Discovery
	transport transport.Transport
}

func NewBridge(opt *Option) *Bridge {
	b := &Bridge{option: opt}

	b.discovery = discovery.NewSerf(&discovery.Opt{
		Addr:      b.option.Addr,
		Advertise: b.option.Advertise,
		Name:      b.option.Name,
		Members:   b.option.Agents,
		PipePort:  b.option.PipePort,
	})
	b.transport = transport.NewRpcTransport(&transport.Opt{
		Port: b.option.PipePort,
	})
	b.transport.SetHandler(b)
	b.discovery.SetHandler(b)
	return b
}

func (b *Bridge) Serve() error {
	if b.option.Broker == nil {
		return ErrInvalidBroker
	}
	b.discovery.Start()
	b.transport.Start()
	return nil
}

func (b *Bridge) OnAgentJoin(a *discovery.Agent) {
	if b.transport != nil {
		b.transport.Join(a)
	}
}

func (b *Bridge) OnAgentLeave(a *discovery.Agent) {
	if b.transport != nil {
		b.transport.Leave(a)
	}
}

func (b *Bridge) OnAgentUpdate(a *discovery.Agent) {
	if b.transport != nil {
		b.transport.Update(a)
	}
}

func (b *Bridge) PushConnect(node *discovery.Agent, clientId string) {
	if b.transport != nil {
		b.transport.PushConnect(node, clientId)
	}
}

func (b *Bridge) PushDisconnect(node *discovery.Agent, clientId string) {
	if b.transport != nil {
		b.transport.PushDisconnect(node, clientId)
	}
}

func (b *Bridge) PushPublish(node *discovery.Agent, topic string, payload []byte, qos byte, retain bool) {
	if b.transport != nil {
		b.transport.PushPublish(node, topic, payload, qos, retain)
	}
}

func (b *Bridge) OnConnect(id string, clientId string) {
	log.Printf("[INFO] client id:%s connected from bridge agent:%s \n", clientId, id)
	if existing, ok := b.option.Broker.Clients.Get(clientId); ok {
		b.option.Broker.DisconnectClient(existing, packets.ErrSessionTakenOver)
	}
}

func (b *Bridge) OnDisConnect(id string, clientId string) {
	log.Printf("[INFO] client id:%s disconnected from bridge agent:%s \n", clientId, id)
}

func (b *Bridge) OnPublish(id string, topic string, payload []byte, qos byte, retain bool) {
	cl := b.option.Broker.NewClient(nil, "local", HookId, true)
	b.option.Broker.InjectPacket(cl, packets.Packet{
		FixedHeader: packets.FixedHeader{
			Type:   packets.Publish,
			Qos:    qos,
			Retain: retain,
		},
		TopicName: topic,
		Payload:   payload,
		PacketID:  uint16(qos),
	})
}
