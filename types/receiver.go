package types

import (
	"context"
	"errors"
	"fmt"
)

type Redirect struct {
	*ErrorChannel

	ctx       context.Context
	cancel    context.CancelFunc
	onMessage func(interface{})
	IsStarted bool

	in  chan interface{}
	out map[ChanName]chan interface{}
}

func NewRedirect(ctx context.Context, errChan *ErrorChannel, from chan interface{}, to map[ChanName]chan interface{}, onMsg func(interface{})) *Redirect {
	return &Redirect{
		in:           from,
		out:          to,
		ctx:          ctx,
		onMessage:    onMsg,
		ErrorChannel: errChan,
	}
}

func (r *Redirect) Start() {
	if r.IsStarted {
		r.SendError(fmt.Errorf("\"%T\" already started", *r))
		return
	}

	ctx, cancel := context.WithCancel(r.ctx)
	r.ctx = ctx
	r.cancel = cancel
	r.IsStarted = true

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

func (r *Redirect) Stop() {
	if !r.IsStarted {
		r.SendError(fmt.Errorf("\"%T\" already stopped", *r))
		return
	}
	r.cancel()
}

func (r *Redirect) Close() {
	if r.IsStarted {
		r.Stop()
	}

	close(r.in)

	for i := range r.out {
		close(r.out[i])
	}
}

func (r *Redirect) Send(data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			r.SendError(err.(error))
		}
	}()

	r.in <- data
}

func (r *Redirect) AppendListeners(listeners map[ChanName]chan interface{}) {
	if r.IsStarted {
		r.Stop()
		defer r.Start()
	}

	for name := range listeners {
		if _, ok := r.out[name]; ok {
			r.SendError(fmt.Errorf("listener with name \"%s\" already exists", name))
		} else {
			r.out[name] = listeners[name]
		}
	}
}

func (r *Redirect) InputChan() chan interface{} {
	return r.in
}

func (r *Redirect) OutputChan() map[ChanName]chan interface{} {
	return r.out
}
