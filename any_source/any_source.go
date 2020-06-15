package any_source

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/broadcast"
	"github.com/KlyuchnikovV/broadcast/types"
)

type AnySource struct {
	*types.Redirect

	redirectors []*broadcast.Broadcast
}

func New(ctx context.Context, errChan *types.ErrorChannel, from []chan interface{}, to chan interface{}) *AnySource {
	if len(to) < 1 {
		return nil
	}

	a := new(AnySource)
	a.redirectors = make([]*broadcast.Broadcast, len(from))

	inChan := make(chan interface{}, 2*len(from))

	for i := range from {
		a.redirectors[i] = broadcast.New(ctx, errChan, from[i], inChan)
	}

	a.Redirect = types.NewRedirect(ctx, errChan, inChan, map[types.ChanName]chan interface{}{"0": to}, a.onMessage)

	return a
}

func (a *AnySource) onMessage(data interface{}) {
	for _, ch := range a.OutputChan() {
		ch <- data
	}
}

func (a *AnySource) Start() {
	a.Redirect.Start()
	for i := range a.redirectors {
		a.redirectors[i].Start()
	}
}

func (a *AnySource) Stop() {
	if !a.IsStarted {
		a.SendError(fmt.Errorf("\"%T\" already stopped", *a))
		return
	}
	for i := range a.redirectors {
		a.redirectors[i].Stop()
	}
	a.Redirect.Stop()
}

func (a *AnySource) Close() {
	if a.IsStarted {
		a.Stop()
	}
	
	for i := range a.redirectors {
		a.redirectors[i].Close()
	}

	a.Redirect.Close()
}

func (a *AnySource) AppendListeners(_ ...chan interface{}) {
	a.SendError(fmt.Errorf("cannot append to \"%T\" listener (not allowed by design)", *a))
}