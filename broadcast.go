package broadcast

import (
	"context"
	"errors"
)

type Broadcast struct {
	broadcaster

	in  chan Message
	out []chan Message
}

func NewBroadcast(ctx context.Context, errChan chan error, from chan Message, to ...chan Message) *Broadcast {
	if len(to) < 1 {
		return nil
	}

	return &Broadcast{
		broadcaster: *newBroadcaster(ctx, errChan),
		in:          from,
		out:         to,
	}
}

func (b *Broadcast) Start() {
	b.broadcaster.Start(func() {
		var msg Message
		for {
			select {
			case msg = <-b.in:
				for _, ch := range b.out {
					ch <- msg
				}
			case <-b.ctx.Done():
				err := b.ctx.Err()
				if err != nil && !errors.Is(err, context.Canceled) {
					b.err <- err
				}
				return
			}
		}
	})
}
