package transport

import "github.com/werbenhu/bridgemq/agent"

type Handler interface {
	OnConnect(id string, clientId string)
	OnDisConnect(id string, clientId string)
	OnPublish(id string, topic string, payload []byte, qos byte, retain bool)
}

type Transport interface {
	Join(node *agent.Agent)
	Leave(node *agent.Agent)
	Update(node *agent.Agent)
	SetHandler(Handler)
	PushConnect(local *agent.Agent, clientId string)
	PushDisconnect(local *agent.Agent, clientId string)
	PushPublish(local *agent.Agent, topic string, payload []byte, qos byte, retain bool)
	Start() error
	Stop()
}
