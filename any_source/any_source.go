package any_source

import (
	"context"
	"fmt"

	"github.com/KlyuchnikovV/broadcast"
	"github.com/KlyuchnikovV/broadcast/types"
)

type AnySource struct {
	types.ErrorChannel
	redirectors map[types.ChanName]*broadcast.Broadcast
}

func New(ctx context.Context, errChan types.ErrorChannel, bufferSizes map[types.ChanName]int, to types.Listener) *AnySource {
	redirectors := make(map[types.ChanName]*broadcast.Broadcast, len(bufferSizes))

	for i, cap := range bufferSizes {
		redirectors[i] = broadcast.New(ctx, errChan, cap, to)
	}

	return &AnySource{
		redirectors:  redirectors,
		ErrorChannel: errChan,
	}
}

func (a *AnySource) Start() {
	for i := range a.redirectors {
		a.redirectors[i].Start()
	}
}

func (a *AnySource) Stop() {
	for i := range a.redirectors {
		a.redirectors[i].Stop()
	}
}

func (a *AnySource) Close() {
	for i := range a.redirectors {
		a.redirectors[i].Close()
	}
}

func (a *AnySource) GetChannels(names... types.ChanName) map[types.ChanName]chan interface{} {
	var result map[types.ChanName]chan interface{}
	
	if len(names) == 0 {
		for name, b := range a.redirectors {
			result[name] = b.GetChan()
		}
	} else {
		for _, name := range names {
			b, ok := a.redirectors[name]
			if !ok {
				a.SendError(fmt.Errorf("channel \"%s\" not found", name))
				continue
			}
			result[name] = b.GetChan()
		}
	}

	return result
}

func (a *AnySource) Send(name types.ChanName, data interface{}) {
	b, ok := a.redirectors[name]
	if !ok {
		a.SendError(fmt.Errorf("channel \"%s\" not found", name))
		return
	}
	b.Send(data)
}
