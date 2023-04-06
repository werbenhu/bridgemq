package bridgemq

import (
	"log"

	"github.com/mochi-co/mqtt/v2/packets"
	"github.com/werbenhu/bridgemq/agent"
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

func (b *Bridge) Stop() error {
	b.transport.Stop()
	b.discovery.Stop()
	return nil
}

func (b *Bridge) LocalAgent() *agent.Agent {
	if b.discovery != nil {
		b.discovery.LocalAgent()
	}
	return nil
}

func (b *Bridge) OnAgentJoin(a *agent.Agent) {
	if b.transport != nil {
		b.transport.Join(a)
	}
}

func (b *Bridge) OnAgentLeave(a *agent.Agent) {
	if b.transport != nil {
		b.transport.Leave(a)
	}
}

func (b *Bridge) OnAgentUpdate(a *agent.Agent) {
	if b.transport != nil {
		b.transport.Update(a)
	}
}

func (b *Bridge) PushConnect(clientId string) {
	if b.transport != nil {
		local := b.discovery.LocalAgent()
		b.transport.PushConnect(local, clientId)
	}
}

func (b *Bridge) PushDisconnect(clientId string) {
	if b.transport != nil {
		local := b.discovery.LocalAgent()
		b.transport.PushDisconnect(local, clientId)
	}
}

func (b *Bridge) PushPublish(topic string, payload []byte, qos byte, retain bool) {
	if b.transport != nil {
		local := b.discovery.LocalAgent()
		b.transport.PushPublish(local, topic, payload, qos, retain)
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
