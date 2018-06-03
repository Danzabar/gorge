package test

import (
	"fmt"
	"testing"

	"github.com/Danzabar/gorge/engine"
	"github.com/stretchr/testify/assert"
)

type (
	TestEvents struct {
		engine.Component
	}
)

func (t *TestEvents) Register() {
	t.Event("test.direct", "", false, []string{engine.DIRECT_CHAN})
	t.Event("test.internal", "", false, []string{engine.INTERNAL_CHAN})
}

func TestEventFiresOnConnection(t *testing.T) {
	app := NewApplicationTest("test-122")
	done := make(chan bool)

	app.GM.RegisterHandler("connected", func(e engine.Event) bool {
		c := e.Data.(*engine.Client)
		assert.Equal(t, app.Client.Id, c.Id)

		done <- true
		return true
	})

	app.Start()
	<-done
}

func TestDirectEvent(t *testing.T) {
	app := NewApplicationTest("test-123")

	go func() {
		app.GM.FireEvent(engine.NewDirectEvent("test.direct", "test-123", ""))
	}()

	go func() {
	eventloop:
		for {
			select {
			case e, _ := <-app.Connection.In:
				fmt.Println(e)
				break eventloop
			}
		}

	}()
}
