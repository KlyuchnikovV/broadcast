package types

import (
	"context"
	"errors"
	"fmt"
)

type Receiver struct {
	*ErrorChannel

	ctx       context.Context
	cancel    context.CancelFunc
	onMessage func(interface{})
	isStarted bool

	in chan interface{}
}

func NewReceiver(ctx context.Context, errChan *ErrorChannel, from chan interface{}, onMsg func(interface{})) *Receiver {
	return &Receiver{
		in:           from,
		ctx:          ctx,
		onMessage:    onMsg,
		ErrorChannel: errChan,
	}
}

func (r *Receiver) Start() {
	if r.isStarted {
		r.SendError(fmt.Errorf("\"%T\" already started", *r))
		return
	}

	ctx, cancel := context.WithCancel(r.ctx)
	r.ctx = ctx
	r.cancel = cancel
	r.isStarted = true

	go func() {
		defer func() {
			if data := recover(); data != nil {
				r.SendError(data.(error))
			}
		}()

		for {
			select {
			case msg, ok := <-r.in:
				if !ok {
					r.SendError(fmt.Errorf("input channel was closed before \"%T\" stopped listening", *r))
					return
				}
				r.onMessage(msg)
			case <-r.ctx.Done():
				err := r.ctx.Err()
				if err != nil && !errors.Is(err, context.Canceled) {
					r.SendError(err)
				}
				return
			}
		}
	}()
}

func (r *Receiver) Stop() {
	if !r.isStarted {
		r.SendError(fmt.Errorf("\"%T\" already stopped", *r))
		return
	}
	r.cancel()
}

func (r *Receiver) Close() {
	if r.isStarted {
		r.Stop()
	}
	close(r.in)
}
