package types

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/chan_utils"
)

type Redirect struct {
	*ErrorChannel

	ctx       context.Context
	cancel    context.CancelFunc
	onMessage func(interface{})
	IsStarted bool

	in  chan interface{}
	out map[ChanName]Listener
}

func NewRedirect(ctx context.Context, errChan *ErrorChannel, from chan interface{}, to map[ChanName]Listener, onMessage func(interface{})) *Redirect {
	return &Redirect{
		in:           from,
		out:          to,
		ctx:          ctx,
		onMessage:    onMessage,
		ErrorChannel: errChan,
	}
}

func (r *Redirect) Start() {
	if r.IsStarted {
		r.SendError(fmt.Errorf("\"%T\" already started", *r))
		return
	}

	if r.in == nil {
		return
	}

	redirect, cancel := chan_utils.NewListener(r.ctx, r.in, r.onMessage, r.SendError)

	r.cancel = cancel
	r.IsStarted = true

	go redirect()
}

func (r *Redirect) Stop() {
	if !r.IsStarted {
		r.SendError(fmt.Errorf("\"%T\" already stopped", *r))
		return
	}
	if r.in == nil {
		return
	}
	r.cancel()
}

func (r *Redirect) Close() {
	if r.IsStarted {
		r.Stop()
	}
	
	if r.in == nil {
		return
	}

	close(r.in)
}

func (r *Redirect) Send(data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			r.SendError(err.(error))
		}
	}()

	r.in <- data
}

func (r *Redirect) AppendListeners(listeners map[ChanName]Listener) {
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

func (r *Redirect) Channel() chan interface{} {
	return r.in
}

func (r *Redirect) Listeners() map[ChanName]Listener {
	return r.out
}
