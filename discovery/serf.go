package discovery

import (
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/logutils"
	"github.com/hashicorp/serf/serf"
	"github.com/natefinch/lumberjack"
	"github.com/werbenhu/bridgemq/agent"
)

const (
	PortKey = "pipe_port"
)

type Serf struct {
	events  chan serf.Event
	opts    *Opt
	serf    *serf.Serf
	handler Handler
	agents  sync.Map
}

type Opt struct {
	Addr      string
	Advertise string
	Name      string
	Members   string
	PipePort  string
}

func NewSerf(opts *Opt) *Serf {
	s := &Serf{
		events: make(chan serf.Event),
		opts:   opts,
	}
	return s
}

func (s *Serf) LocalAgent() *agent.Agent {
	node, ok := s.agents.Load(s.opts.Name)
	if !ok {
		return nil
	}
	return node.(*agent.Agent)
}

func (s *Serf) Agents() []*agent.Agent {
	nodes := make([]*agent.Agent, 0)
	s.agents.Range(func(key any, val any) bool {
		nodes = append(nodes, val.(*agent.Agent))
		return true
	})
	return nodes
}

func (s *Serf) SetHandler(h Handler) {
	s.handler = h
}

func (s *Serf) Stop() {
	s.serf.Shutdown()
	close(s.events)
}

func (s *Serf) Start() error {
	var err error
	cfg := serf.DefaultConfig()
	cfg.MemberlistConfig.AdvertiseAddr, cfg.MemberlistConfig.AdvertisePort = s.splitHostPort(s.opts.Advertise)
	cfg.MemberlistConfig.BindAddr, cfg.MemberlistConfig.BindPort = s.splitHostPort(s.opts.Addr)
	cfg.EventCh = s.events

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer: io.MultiWriter(&lumberjack.Logger{
			Filename:   "./log/serf.log",
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
		}, os.Stderr),
	}

	cfg.Logger = log.New(os.Stderr, "", log.LstdFlags)
	cfg.Logger.SetOutput(filter)
	cfg.MemberlistConfig.Logger = cfg.Logger
	cfg.NodeName = s.opts.Name

	s.serf, err = serf.Create(cfg)
	if err != nil {
		return err
	}

	s.serf.SetTags(map[string]string{
		PortKey: s.opts.PipePort,
	})
	go s.Loop()
	log.Printf("[INFO] serf discovery started, current agent addr:%s, advertise addr:%s\n", s.opts.Addr, s.opts.Advertise)
	if len(s.opts.Members) > 0 {
		members := strings.Split(s.opts.Members, ",")
		s.Join(members)
	}
	return nil
}

func (s *Serf) Join(members []string) error {
	_, err := s.serf.Join(members, true)
	return err
}

func (s *Serf) splitHostPort(addr string) (string, int) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatalf("[ERROR] serf discovery parse addr:%s err:%s", addr, err.Error())
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		log.Fatalf("[ERROR] serf discovery parse port:%s err:%s", p, err.Error())
	}
	return h, port
}

func (s *Serf) Loop() {
	for e := range s.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				node := agent.New(member.Name, member.Addr.String(), member.Port, member.Tags[PortKey])
				if s.opts.Name != member.Name {
					s.handler.OnAgentJoin(node)
				}
				s.agents.Store(node.Id, node)
			}

		case serf.EventMemberUpdate:
			for _, member := range e.(serf.MemberEvent).Members {
				node := agent.New(member.Name, member.Addr.String(), member.Port, member.Tags[PortKey])
				if s.serf.LocalMember().Name != member.Name {
					s.handler.OnAgentUpdate(node)
				}
				s.agents.Store(node.Id, node)
			}

		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				node := agent.New(member.Name, member.Addr.String(), member.Port, member.Tags[PortKey])
				if s.serf.LocalMember().Name == member.Name {
					s.handler.OnAgentLeave(node)
					s.agents.Delete(node.Id)
				}
			}
		}
	}
}
