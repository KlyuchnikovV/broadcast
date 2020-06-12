package broadcast

import (
	"context"
	"fmt"
	"github.com/KlyuchnikovV/broadcast/types"
)

type Broadcast struct {
	types.Receiver

	out []chan interface{}
}

func New(ctx context.Context, errChan *types.ErrorChannel, from chan interface{}, to ...chan interface{}) *Broadcast {
	if len(to) < 1 {
		return nil
	}

	b := &Broadcast{out: to}

	b.Receiver = *types.NewReceiver(ctx, errChan, from, b.onMessage)

	return b
}

func (b *Broadcast) onMessage(data interface{}) {
	msg, ok := data.(types.Message)
	if !ok {
		b.SendError(fmt.Errorf("message \"%v\" wasn't of type \"%T\"", data, msg))
		return
	}

	for _, ch := range b.out {
		ch <- msg
	}
}

func (b *Broadcast) Close() {
	b.Receiver.Close()

	for _, ch := range b.out {
		close(ch)
	}
}

func (b *Broadcast) AppendListeners(listeners ...chan interface{}) {
	if b.IsStarted {
		b.Stop()
		defer b.Start()
	}

	b.out = append(b.out, listeners...)
}