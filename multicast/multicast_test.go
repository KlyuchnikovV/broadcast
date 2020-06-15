package multicast

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/KlyuchnikovV/broadcast/types"
	"github.com/stretchr/testify/assert"
)

type DirectMessage struct {
	msg   interface{}
	names []types.ChanName
}

func (d *DirectMessage) GetNames() []types.ChanName {
	return d.names
}

func initTestMulticast(t *testing.T, outN int) (*Multicast, chan error, context.CancelFunc, *sync.WaitGroup) {
	var input = make(chan interface{})
	var outputs = make(map[types.ChanName]chan interface{})
	var err = make(chan error, 1000)
	var errChan = types.NewErrorChannel(err, &sync.Mutex{})

	for i := 0; i < outN; i++ {
		outputs[types.ChanName(strconv.Itoa(i))] = make(chan interface{})
	}

	ctx, cancel := context.WithCancel(context.Background())

	b := New(ctx, errChan, input, outputs)

	wg := new(sync.WaitGroup)

	for i := range outputs {
		go func(name types.ChanName, ch chan interface{}) {
			for i := 0; ; i++ {
				select {
				case message, ok := <-ch:
					assert.True(t, ok)
					fmt.Printf("\"%s\" got message \"%d\"\n", name, message.(*DirectMessage).msg.(int))
					wg.Done()
				case <-ctx.Done():
					return
				}
			}
		}(i, outputs[i])
	}
	return b, err, cancel, wg
}

func TestMulticast(t *testing.T) {
	rand.Seed(1)

	var outN = 10
	b, err, cancel, wg := initTestMulticast(t, outN)
	defer cancel()
	defer assert.Empty(t, err)

	b.Start()

	for i := 0; i < 1000; i++ {
		var names []types.ChanName
		for i := rand.Intn(outN); i > 0; i-- {
			names = append(names, types.ChanName(strconv.Itoa(i)))
		}

		numberOfWaits := outN
		if len(names) > 0 {
			numberOfWaits = len(names)
		}

		wg.Add(numberOfWaits)
		b.Send(&DirectMessage{
			msg:   i,
			names: names,
		})
		wg.Wait()
		assert.Empty(t, err)
	}

	b.Stop()
	assert.Empty(t, err)
}

// TODO: concurrency test
// TODO: benchmarks
