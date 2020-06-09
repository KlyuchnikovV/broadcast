package broadcast

import (
	"context"
	"fmt"
)

type ChanName string

// TODO: write interface
type Message interface {
}

type NamedMessage interface {
	Message
	GetNames() []ChanName
}

type broadcaster struct {
	ctx       context.Context
	cancel    context.CancelFunc
	err       chan error
	isStarted bool
}

func newBroadcaster(ctx context.Context, errChan chan error) *broadcaster {
	ctx, cancel := context.WithCancel(ctx)
	return &broadcaster{
		ctx:    ctx,
		cancel: cancel,
		err:    errChan,
	}
}

func (b *broadcaster) Start(broadcast func()) {
	if b.isStarted {
		b.err <- fmt.Errorf("broadcaster already started")
		return
	} else {
		b.isStarted = true
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.err <- r.(error)
			}
		}()

		broadcast()
	}()
}

func (b *broadcaster) Stop() {
	b.cancel()
}
