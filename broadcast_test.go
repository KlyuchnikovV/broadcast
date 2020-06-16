package broadcast

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/KlyuchnikovV/broadcast/types"
	"github.com/stretchr/testify/assert"
)

func initTestBroadcast(t *testing.T, outN int, withWg bool) (*Broadcast, chan error, context.CancelFunc, *sync.WaitGroup) {
	var input = make(chan interface{})
	var outputs = make([]chan interface{}, outN)
	var err = make(chan error, 1000)
	var errChan = types.NewErrorChannel(err, &sync.Mutex{})

	for i := range outputs {
		outputs[i] = make(chan interface{})
	}

	ctx, cancel := context.WithCancel(context.Background())

	b := New(ctx, errChan, input, outputs...)

	wg := new(sync.WaitGroup)

	for i := range outputs {
		go func(n int, ch chan interface{}) {
			for i := 0; ; i++ {
				select {
				case message, ok := <-ch:
					assert.True(t, ok)
					log.Printf("\"%d\" got message \"%d\"\n", n, message.(int))
					if withWg {
						wg.Done()
					}
				case <-ctx.Done():
					return
				}
			}
		}(i, outputs[i])
	}
	return b, err, cancel, wg
}

func TestBroadcast(t *testing.T) {
	var outN = 10
	b, err, cancel, wg := initTestBroadcast(t, outN, true)
	defer cancel()

	b.Start()

	for i := 0; i < 1000; i++ {
		wg.Add(outN)
		b.Send(i)
		wg.Wait()
		assert.Empty(t, err)
	}

	b.Stop()
	assert.Empty(t, err)
}

func TestBroadcastConcurrency(t *testing.T) {
	var outN = 10
	b, err, cancel, _ := initTestBroadcast(t, outN, false)
	defer cancel()

	b.Start()

	inWg := new(sync.WaitGroup)

	inWg.Add(outN)
	for i := 0; i < outN; i++ {
		go func(i int) {
			inWg.Done()
			inWg.Wait()
			b.Send(i)
			log.Printf("Sended %d\n", i)
		}(i)
		assert.Empty(t, err)
	}

	assert.Empty(t, err)
	b.Stop()
}

// TODO: benchmarks
