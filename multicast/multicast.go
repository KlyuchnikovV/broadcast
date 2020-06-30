package multicast

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/broadcast/types"
)

type Multicast struct {
	types.Receiver

	listeners map[types.ChanName]types.Listener
}

func New(ctx context.Context, errChan types.ErrorChannel, bufferSize int, to map[types.ChanName]types.Listener) *Multicast {
	if len(to) < 1 {
		return nil
	}

	return &Multicast{
		Receiver:  *types.NewReceiver(ctx, errChan, types.ChannelCapacity(bufferSize)),
		listeners: to,
	}
}

func (m *Multicast) Start() {
	m.Receiver.Start(m.Send)
}

func (m *Multicast) Send(data interface{}) {
	msg, ok := data.(types.DirectedMessage)
	if !ok {
		m.SendError(fmt.Errorf("message \"%v\" wasn't of type \"%T\"", data, msg))
		return
	}

	names := msg.GetNames()

	if len(names) == 0 {
		for _, listener := range m.listeners {
			listener(data)
		}
		return
	}

	for _, name := range names {
		if listener, ok := m.listeners[name]; ok {
			listener(data)
			continue
		}

		m.SendError(fmt.Errorf("channel \"%s\" not found (message was: %#v)", name, msg))
	}
}
