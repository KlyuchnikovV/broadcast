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
	e.mutex.Lock()
	e.err <- err
	e.mutex.Unlock()
}

func (e *ErrorChannel) Close() {
	e.mutex.Lock()
	close(e.err)
	e.mutex.Unlock()
}
