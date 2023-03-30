package bridgemq

import (
	"bytes"
	"log"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

const (
	HookId = "bridgehook"
)

// Hook is a debugging hook which logs additional low-level information from the server.
type Hook struct {
	bridge *Bridge
	mqtt.HookBase
}

func (h *Hook) ID() string {
	return HookId
}

// Provides indicates which hook methods this hook provides.
func (h *Hook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnSessionEstablished,
		mqtt.OnDisconnect,
		mqtt.OnWillSent,
		mqtt.OnPublish,
	}, []byte{b})
}

// Init is called when the hook is initialized.
func (h *Hook) Init(config any) error {
	if _, ok := config.([]IOption); !ok && config != nil {
		return mqtt.ErrInvalidConfigType
	}

	option := DefaultOption()
	for _, o := range config.([]IOption) {
		o(option)
	}

	h.bridge = NewBridge(option)
	go h.bridge.Serve()
	return nil
}

func (h *Hook) OnSessionEstablished(cl *mqtt.Client, pk packets.Packet) {
	log.Printf("[INFO] local client id:%s connected \n", cl.ID)
	h.bridge.PushConnect(h.bridge.discovery.LocalAgent(), pk.Connect.ClientIdentifier)
}

func (h *Hook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {

	if cl.ID == h.ID() && cl.Net.Inline {
		return pk, nil
	}
	h.pushPublish(cl, pk)
	return pk, nil
}

func (h *Hook) OnWillSent(cl *mqtt.Client, pk packets.Packet) {
	h.pushPublish(cl, pk)
}

func (h *Hook) OnDisconnect(cl *mqtt.Client, err error, expire bool) {
	log.Printf("[INFO] local client id:%s disconnected \n", cl.ID)
	h.bridge.PushDisconnect(h.bridge.discovery.LocalAgent(), cl.ID)
}

func (h *Hook) pushPublish(cl *mqtt.Client, pk packets.Packet) {
	h.bridge.PushPublish(
		h.bridge.discovery.LocalAgent(),
		pk.TopicName,
		pk.Payload,
		pk.FixedHeader.Qos,
		pk.FixedHeader.Retain,
	)
}

// Stop is called to gracefully shutdown the hook.
func (h *Hook) Stop() error {
	return h.bridge.Stop()
}
