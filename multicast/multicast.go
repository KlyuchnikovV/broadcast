package multicast

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/broadcast/types"
)

type Multicast struct {
	types.Redirect
}

func New(ctx context.Context, errChan *types.ErrorChannel, from chan interface{}, to map[types.ChanName]types.Listener) *Multicast {
	if len(to) < 1 {
		return nil
	}

	m := new(Multicast)

	m.Redirect = *types.NewRedirect(ctx, errChan, from, to, m.OnMessage)

	return m
}

func NewWithoutChan(ctx context.Context, errChan *types.ErrorChannel, to map[types.ChanName]types.Listener) *Multicast {
	if len(to) < 1 {
		return nil
	}

	b := new(Multicast)
	b.Redirect = *types.NewRedirect(ctx, errChan, nil, to, b.OnMessage)

	return b
}

func (m *Multicast) OnMessage(data interface{}) {
	msg, ok := data.(types.DirectedMessage)
	if !ok {
		m.SendError(fmt.Errorf("message \"%v\" wasn't of type \"%T\"", data, msg))
		return
	}

	out := m.Listeners()

	names := msg.GetNames()
	if len(names) != 0 {
		for _, name := range names {
			if listener, ok := out[name]; ok {
				listener.OnMessage(data)
			} else {
				m.SendError(fmt.Errorf("channel \"%s\" not found (message was: %#v)", name, msg))
			}
		}
	} else {
		// No names provided -> sending to all
		for _, listener := range out {
			listener.OnMessage(data)
		}
	}
}
