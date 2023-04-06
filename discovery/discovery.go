package discovery

import "github.com/werbenhu/bridgemq/agent"

type Handler interface {
	OnAgentJoin(*agent.Agent)
	OnAgentLeave(*agent.Agent)
	OnAgentUpdate(*agent.Agent)
}

type Discovery interface {
	SetHandler(Handler)
	Agents() []*agent.Agent
	LocalAgent() *agent.Agent
	Start() error
	Stop()
}
