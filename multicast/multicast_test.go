package multicast

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/KlyuchnikovV/broadcast/types"
	"github.com/stretchr/testify/assert"
)

var cancel context.CancelFunc

type simpleListener int

func (s simpleListener) OnMessage(data interface{}) {
	msg, ok := data.(*directMessage)
	if !ok {
		cancel()
		panic(fmt.Errorf("message not of type %T", msg))
	}
	log.Printf("\"%d\" got message \"%d\"\n",s, msg.data.(int))
	if msg.withWg {
		msg.wg.Done()
	}
} 

type directMessage struct {
	wg     *sync.WaitGroup
	withWg bool
	data   interface{}
	names []types.ChanName
}

func (d *directMessage) GetNames() []types.ChanName {
	return d.names
}

func initTestMulticast(t *testing.T, outN int) (*Multicast, *types.ErrorChannel, *sync.WaitGroup) {
	var outputs = make(map[types.ChanName]types.Listener)
	var errChan = types.NewErrorChannel(1)

	for i := 0; i < outN; i++ {
		outputs[types.ChanName(strconv.Itoa(i))] = simpleListener(i).OnMessage
	}

	ctx, cancellation := context.WithCancel(context.Background())
	cancel = cancellation

	b := New(ctx, errChan, 1, outputs)

	wg := new(sync.WaitGroup)
	return b, errChan, wg
}

func TestMulticast(t *testing.T) {
	rand.Seed(1)

	var outN = 10
	b, err, wg := initTestMulticast(t, outN)
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
		b.Send(&directMessage{
			data: i,
			wg: wg,
			withWg: true,
			names: names,
		})
		wg.Wait()
		assert.Empty(t, err)
	}

	b.Stop()
	assert.Empty(t, err)
}
