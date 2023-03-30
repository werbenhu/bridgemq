package bridge

import (
	"os"

	"github.com/mochi-co/mqtt/v2"
	"github.com/rs/xid"
)

type Option struct {
	Name      string
	Addr      string
	Advertise string
	PipePort  string
	Agents    string
	Broker    *mqtt.Server
}

type IOption func(o *Option)

func OptBroker(b *mqtt.Server) IOption {
	return func(o *Option) {
		o.Broker = b
	}
}

func OptPipePort(p string) IOption {
	return func(o *Option) {
		o.PipePort = p
	}
}

func OptName(name string) IOption {
	return func(o *Option) {
		if name != "" {
			o.Name = name
		}
	}
}

func OptAdvertise(addr string) IOption {
	return func(o *Option) {
		if addr != "" {
			o.Advertise = addr
		}
	}
}

func OptAddr(addr string) IOption {
	return func(o *Option) {
		if addr != "" {
			o.Addr = addr
		}
	}
}

func OptAgents(agents string) IOption {
	return func(o *Option) {
		if agents != "" {
			o.Agents = agents
		}
	}
}

func DefaultOption() *Option {
	hostname, _ := os.Hostname()
	return &Option{
		Name:      hostname + "-" + xid.New().String(),
		Addr:      ":7933",
		Advertise: ":7933",
		PipePort:  "8933",
	}
}
