package multicast

import (
	"context"
	"fmt"
	"github.com/KlyuchnikovV/broadcast/types"
)

type Multicast struct {
	types.Receiver

	in  chan types.DirectedMessage
	out map[types.ChanName]chan types.Message
}

func New(ctx context.Context, errChan *types.ErrorChannel, from chan interface{}, to map[types.ChanName]chan types.Message) *Multicast {
	if len(to) < 1 {
		return nil
	}

	m := &Multicast{out: to}

	m.Receiver = *types.NewReceiver(ctx, errChan, from, m.onMessage)

	return m
}

func (m *Multicast) onMessage(data interface{}) {
	msg, ok := data.(types.DirectedMessage)
	if !ok {
		m.SendError(fmt.Errorf("message \"%v\" wasn't of type \"%T\"", data, msg))
		return
	}

	names := msg.GetNames()
	if len(names) == 0 {
		// No names provided -> sending to all
		for _, ch := range m.out {
			ch <- msg
		}
		return
	}

	for _, name := range names {
		if ch, ok := m.out[name]; ok {
			ch <- msg
		} else {
			m.SendError(fmt.Errorf("channel \"%m\" not found (message was: %#v)", name, msg))
		}
	}
}

func (m *Multicast) Close() {
	m.Receiver.Close()

	for _, ch := range m.out {
		close(ch)
	}
}
