package any_source

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/broadcast"
	"github.com/KlyuchnikovV/broadcast/types"
)

type AnySource struct {
	*types.Redirect

	redirs []*broadcast.Broadcast
}

func New(ctx context.Context, errChan *types.ErrorChannel, from []chan interface{}, to types.Listener) *AnySource {
	a := new(AnySource)
	a.redirs = make([]*broadcast.Broadcast, len(from))

	for i := range from {
		a.redirs[i] = broadcast.New(ctx, errChan, from[i], to)
	}

	a.Redirect = types.NewRedirect(ctx, errChan, nil, map[types.ChanName]types.Listener{"0": to}, a.OnMessage)

	return a
}

func (a *AnySource) OnMessage(data interface{}) {
	for _, listener := range a.Listeners() {
		listener.OnMessage(data)
	}
}

func (a *AnySource) Start() {
	a.Redirect.Start()
	for i := range a.redirs {
		a.redirs[i].Start()
	}
}

func (a *AnySource) Stop() {
	if !a.IsStarted {
		a.SendError(fmt.Errorf("\"%T\" already stopped", *a))
		return
	}
	for i := range a.redirs {
		a.redirs[i].Stop()
	}
	a.Redirect.Stop()
}

func (a *AnySource) Close() {
	if a.IsStarted {
		a.Stop()
	}

	for i := range a.redirs {
		a.redirs[i].Close()
	}

	a.Redirect.Close()
}

func (a *AnySource) AppendListeners(listeners ...types.Listener) {
	listener, ok := a.Listeners()["0"].(interface {
		AppendListeners(...types.Listener)
	})

	if ok {
		listener.AppendListeners(listeners...)
	} else {
		a.SendError(fmt.Errorf("cannot append to \"%T\" listener (not allowed)", *a))
	}
}
