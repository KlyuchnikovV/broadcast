package types

import "sync"

type ChanName string

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
