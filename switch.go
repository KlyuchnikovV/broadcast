package broadcast

import (
	"context"
	"errors"
	"fmt"
)

type Switch struct {
	broadcaster

	in  chan NamedMessage
	out map[ChanName]chan Message
}

// TODO: several chans for message
func NewSwitch(ctx context.Context, errChan chan error, from chan NamedMessage, to map[ChanName]chan Message) *Switch {
	if len(to) < 1 {
		return nil
	}

	return &Switch{
		broadcaster: *newBroadcaster(ctx, errChan),
		in:          from,
		out:         to,
	}
}

func (s *Switch) Start() {
	s.broadcaster.Start(func() {
		var msg NamedMessage
		for {
			select {
			case msg = <-s.in:
				names := msg.GetNames()
				if len(names) == 0 {
					for _, ch := range s.out {
						ch <- msg
					}
					continue
				}

				for _, name := range names {
					if ch, ok := s.out[name]; ok {
						ch <- msg
					} else {
						s.err <- fmt.Errorf("channel \"%s\" not found (message was: %#v)", name, msg)
					}
				}
			case <-s.ctx.Done():
				err := s.ctx.Err()
				if err != nil && !errors.Is(err, context.Canceled) {
					s.err <- err
				}
				return
			}
		}
	})
}
