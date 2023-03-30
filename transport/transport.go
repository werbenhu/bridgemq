package transport

import (
	"github.com/werbenhu/bridgemq/discovery"
)

type Handler interface {
	OnConnect(id string, clientId string)
	OnDisConnect(id string, clientId string)
	OnPublish(id string, topic string, payload []byte, qos byte, retain bool)
}

type Transport interface {
	Join(node *discovery.Agent)
	Leave(node *discovery.Agent)
	Update(node *discovery.Agent)
	SetHandler(Handler)
	PushConnect(node *discovery.Agent, clientId string)
	PushDisconnect(node *discovery.Agent, clientId string)
	PushPublish(node *discovery.Agent, topic string, payload []byte, qos byte, retain bool)
	Start() error
	Stop()
}
