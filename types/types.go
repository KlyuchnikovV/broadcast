package types

type Listener func(data interface{})

type ChanName string

type DirectedMessage interface {
	GetNames() []ChanName
}

type ErrorChannel chan error

func NewErrorChannel(capacity int) *ErrorChannel {
	var result ErrorChannel

	if capacity < 0 {
		capacity = 0
	}
	
	result = make(chan error, capacity)
	return &result
}

func (e ErrorChannel) SendError(err error) {
	e <- err
}

func (e ErrorChannel) Close() {
	close(e)
}

func (e ErrorChannel) Error() error {
	return <- e
}
