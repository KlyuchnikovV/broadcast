package multicast

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/broadcast/types"
)

type Multicast struct {
	types.Redirect
}

func New(ctx context.Context, errChan *types.ErrorChannel, from chan interface{}, to map[types.ChanName]chan interface{}) *Multicast {
	if len(to) < 1 {
		return nil
	}

	m := new(Multicast)

	m.Redirect = *types.NewRedirect(ctx, errChan, from, to, m.onMessage)

	return m
}

func (m *Multicast) onMessage(data interface{}) {
	msg, ok := data.(types.DirectedMessage)
	if !ok {
		m.SendError(fmt.Errorf("message \"%v\" wasn't of type \"%T\"", data, msg))
		return
	}

	out := m.OutputChan()

	names := msg.GetNames()
	if len(names) == 0 {
		// No names provided -> sending to all
		for _, ch := range out {
			ch <- msg
		}
		return
	}

	for _, name := range names {
		if ch, ok := out[name]; ok {
			ch <- msg
		} else {
			m.SendError(fmt.Errorf("channel \"%s\" not found (message was: %#v)", name, msg))
		}
	}
}
