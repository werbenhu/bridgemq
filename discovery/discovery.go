package discovery

type Agent struct {
	Id       string
	Addr     string
	Port     uint16
	PipePort string
}

type Handler interface {
	OnAgentJoin(*Agent)
	OnAgentLeave(*Agent)
	OnAgentUpdate(*Agent)
}

type Discovery interface {
	SetHandler(Handler)
	Agents() []*Agent
	LocalAgent() *Agent
	Start() error
	Stop()
}
