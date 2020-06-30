package broadcast

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/KlyuchnikovV/broadcast/types"
	"github.com/stretchr/testify/assert"
)

var cancel context.CancelFunc

type simpleListener int

func (s simpleListener) OnMessage(data interface{}) {
	msg, ok := data.(message)
	if !ok {
		cancel()
		panic(fmt.Errorf("message not of type %T", msg))
	}
	log.Printf("\"%d\" got message \"%d\"\n", s, msg.data.(int))
	if msg.withWg {
		msg.wg.Done()
	}
}

type message struct {
	wg     *sync.WaitGroup
	withWg bool
	data   interface{}
}

func initTestBroadcast(t *testing.T, outN int, withWg bool) (*Broadcast, types.ErrorChannel, *sync.WaitGroup) {
	var listeners = make([]types.Listener, outN)
	var errChan = types.NewErrorChannel(1)

	ctx, cancellation := context.WithCancel(context.Background())
	cancel = cancellation

	for i := range listeners {
		listeners[i] = simpleListener(i).OnMessage
	}

	b := New(ctx, errChan, 1, listeners...)
	wg := new(sync.WaitGroup)
	return b, errChan, wg
}

func TestBroadcast(t *testing.T) {
	var outN = 10
	b, err, wg := initTestBroadcast(t, outN, true)
	defer cancel()

	b.Start()

	for i := 0; i < 1000; i++ {
		wg.Add(outN)
		b.Send(message{
			wg:     wg,
			withWg: true,
			data:   i,
		})
		wg.Wait()
		assert.Empty(t, err)
	}

	b.Stop()
	assert.Empty(t, err)
}

func TestBroadcastConcurrency(t *testing.T) {
	var outN = 10
	b, err, _ := initTestBroadcast(t, outN, false)
	defer cancel()

	b.Start()

	inWg := new(sync.WaitGroup)

	inWg.Add(outN)
	for i := 0; i < outN; i++ {
		go func(i int) {
			inWg.Done()
			inWg.Wait()
			b.Send(message{
				data:   i,
			})
			log.Printf("Sended %d\n", i)
		}(i)
		
		assert.Empty(t, err)
	}

	assert.Empty(t, err)
	b.Stop()
}

