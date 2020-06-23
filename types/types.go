package types

import (
	"context"
	"fmt"
	"sync"

	"github.com/KlyuchnikovV/chan_utils"
)

type ChanName string

type Listener interface{
	OnMessage(data interface{})
}

type DirectedMessage interface {
	GetNames() []ChanName
}

type ErrorChannel struct {
	err   chan error
	mutex *sync.Mutex
}

func NewErrorChannel(err chan error, mutex *sync.Mutex) *ErrorChannel {
	return &ErrorChannel{
		err:   err,
		mutex: mutex,
	}
}

func (e *ErrorChannel) SendError(err error) {
	e.err <- err
}

func (e *ErrorChannel) Close() {
	close(e.err)
}

func (e *ErrorChannel) GetError() error {
	return <-e.err
}

type SenderService interface{
	Start(ctx context.Context, onMsg func(interface{}), onErr func(error)) context.CancelFunc
	GetChannel() chan interface{}
	Close()
}

type emptyService chan interface{}

func (emptyService) Start(_ context.Context, _ func(interface{}), _ func(error)) context.CancelFunc {
	return nil
}

func (e emptyService) GetChannel() chan interface{} {
	return e
}

func (emptyService) Close() {
}

type channelService chan interface{}

func (ch channelService) Start(ctx context.Context, onMsg func(interface{}), onErr func(error)) context.CancelFunc {
	redirect, cancel := chan_utils.NewListener(ctx, ch, onMsg, onErr)

	go redirect()

	return cancel
}

func (ch channelService) GetChannel() chan interface{} {
	return ch
}

func (ch channelService) Close() {
	close(ch)
}