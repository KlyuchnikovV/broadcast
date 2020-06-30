package broadcast

import (
	"context"

	"github.com/KlyuchnikovV/broadcast/types"
)

type Broadcast struct {
	types.Receiver
	listeners []types.Listener
}

func New(ctx context.Context, errChan types.ErrorChannel, bufferSize int, to ...types.Listener) *Broadcast {
	if len(to) < 1 {
		return nil
	}

	return &Broadcast{
		Receiver: *types.NewReceiver(ctx, errChan, types.ChannelCapacity(bufferSize)),
		listeners: to,
	}
}

func (b *Broadcast) Start() {
	b.Receiver.Start(b.Send)
}

func (b *Broadcast) Send(data interface{}) {
	for _, listener := range b.listeners {
		listener(data)
	}
}

func (b *Broadcast) AddListener(listener types.Listener) {
	b.listeners = append(b.listeners, listener)
}
