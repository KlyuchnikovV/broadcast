package types

import (
	"context"
	"fmt"
)

type Redirect struct {
	*ErrorChannel

	ctx       context.Context
	cancel    context.CancelFunc
	onMessage func(interface{})
	IsStarted bool

	in  SenderService
	out map[ChanName]Listener
}

func NewRedirect(ctx context.Context, errChan *ErrorChannel, from chan interface{}, to map[ChanName]Listener, onMessage func(interface{})) *Redirect {
	var in SenderService = emptyService(from)

	if from != nil {
		in = channelService(from)
	}

	return &Redirect{
		in:           in,
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

	r.cancel = r.in.Start(r.ctx, r.onMessage, r.SendError)
	r.IsStarted = true
}

func (r *Redirect) Stop() {
	if !r.IsStarted {
		r.SendError(fmt.Errorf("\"%T\" already stopped", *r))
		return
	}
	if r.cancel != nil {
		r.cancel()
	}
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

func (r *Redirect) Listeners() map[ChanName]Listener {
	return r.out
}

func (r *Redirect) Send(data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			r.SendError(err.(error))
		}
	}()

	r.in.GetChannel() <- data
}

func (r *Redirect) GetChannel() chan interface{} {
	return r.in.GetChannel()
}

func (r *Redirect) Close() {
	if r.IsStarted {
		r.Stop()
	}

	if r.in == nil {
		return
	}

	r.in.Close()
}

func (r *Redirect) TransmitTo(name ChanName, listener Listener) {
	r.AppendListeners(map[ChanName]Listener{name: listener})
}
