package broadcast

import (
	"context"
	"strconv"

	"github.com/KlyuchnikovV/broadcast/types"
)

type Broadcast struct {
	types.Redirect
}

func New(ctx context.Context, errChan *types.ErrorChannel, from chan interface{}, to ...types.Listener) *Broadcast {
	if len(to) < 1 {
		return nil
	}

	out := sliceToMap(0, to...)

	b := new(Broadcast)
	b.Redirect = *types.NewRedirect(ctx, errChan, from, out, b.OnMessage)

	return b
}

func (b *Broadcast) OnMessage(data interface{}) {
	for _, listener := range b.Listeners() {
		listener.OnMessage(data)
	}
}

func (b *Broadcast) AppendListeners(listeners ...types.Listener) {
	b.Redirect.AppendListeners(sliceToMap(len(b.Listeners()), listeners...))
}

func sliceToMap(index int, listeners ...types.Listener) map[types.ChanName]types.Listener {
	out := make(map[types.ChanName]types.Listener)

	for i := range listeners {
		out[types.ChanName(strconv.Itoa(index+i))] = listeners[i]
	}

	return out
}
