package any_source

import (
	"context"
	"testing"

	"github.com/KlyuchnikovV/broadcast"
	"github.com/KlyuchnikovV/broadcast/types"
	"github.com/stretchr/testify/assert"
)

func TestAnySourceWithBroadcast(t *testing.T) {
	err := types.NewErrorChannel(1)

	b := broadcast.New(context.Background(), err, -1, func(data interface{}) {
		t.Logf("1: %v", data)
	}, func(data interface{}) {
		t.Logf("2: %v", data)
	})
	a := New(context.Background(), err, map[types.ChanName]int{"a": 1, "b": 1}, b.Send)

	a.Start()
	defer a.Close()

	a.Send("a", "a")
	a.Send("b", "b")

	assert.Empty(t, err)
}
