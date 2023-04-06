package agent

type Agent struct {
	Id       string
	Addr     string
	Port     uint16
	PipePort string
}

func New(id string, addr string, port uint16, pipePort string) *Agent {
	return &Agent{
		Id:       id,
		Addr:     addr,
		Port:     port,
		PipePort: pipePort,
	}
}

func (a *Agent) IsSelf(id string) bool {
	return id == a.Id
}
