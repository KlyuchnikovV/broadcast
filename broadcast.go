package broadcast

import (
	"context"
	"strconv"

	"github.com/KlyuchnikovV/broadcast/types"
)

type Broadcast struct {
	types.Redirect
}

func New(ctx context.Context, errChan *types.ErrorChannel, from chan interface{}, to ...chan interface{}) *Broadcast {
	if len(to) < 1 {
		return nil
	}

	out := sliceToMap(0, to...)

	b := new(Broadcast)
	b.Redirect = *types.NewRedirect(ctx, errChan, from, out, b.onMessage)

	return b
}

func (b *Broadcast) onMessage(data interface{}) {
	for _, ch := range b.OutputChan() {
		ch <- data
	}
}

func (b *Broadcast) AppendListeners(listeners ...chan interface{}) {
	b.Redirect.AppendListeners(sliceToMap(len(b.OutputChan()), listeners...))
}

func sliceToMap(index int, channels ...chan interface{}) map[types.ChanName]chan interface{} {
	out := make(map[types.ChanName]chan interface{})

	for i := range channels {
		out[types.ChanName(strconv.Itoa(index+i))] = channels[i]
	}

	return out
}
