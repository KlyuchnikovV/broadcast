package types

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/chan_utils"
)

type ChannelCapacity int

const NoChannel ChannelCapacity = -1

type Receiver struct {
	ErrorChannel

	ctx context.Context
	// Also acts as IsStarted flag
	cancel context.CancelFunc

	in chan interface{}
}

func NewReceiver(ctx context.Context, errChan ErrorChannel, inChanCapacity ChannelCapacity) *Receiver {
	var in chan interface{} = nil
	if inChanCapacity >= 0 {
		in = make(chan interface{}, inChanCapacity)
	}

	return &Receiver{
		in:           in,
		ctx:          ctx,
		ErrorChannel: errChan,
	}
}

func (r *Receiver) Start(onMessage func(interface{})) {
	if r.IsStarted() {
		r.SendError(fmt.Errorf("\"%T\" already started", *r))
		return
	}

	receive, cancel := chan_utils.NewListener(r.ctx, r.in, onMessage, r.SendError)

	go receive()

	r.cancel = cancel
}

func (r *Receiver) Stop() {
	if !r.IsStarted() {
		r.SendError(fmt.Errorf("\"%T\" already stopped", *r))
		return
	}

	r.cancel()
	r.cancel = nil
}

func (r *Receiver) Close() {
	if r.IsStarted() {
		r.Stop()
	}
	close(r.in)
}

func (r *Receiver) Receive() interface{} {
	if r.in != nil {
		return <-r.in
	}
	return nil
}

func (r *Receiver) Send(data interface{}) {
	if r.in != nil {
		r.in <- data
	}
}

func (r *Receiver) IsStarted() bool {
	return r.cancel != nil
}

func (r *Receiver) GetChan() chan interface{} {
	return r.in
}
