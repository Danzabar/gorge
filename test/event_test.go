package test

import (
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
	t.Event("test.direct", []string{engine.DirectChan})
	t.Event("test.internal", []string{engine.InternalChan})
}

func TestEventFiresOnConnection(t *testing.T) {
	app := NewApplicationTest("test-122")
	done := make(chan bool)

	app.GM.RegisterHandler("connected", func(e engine.Event) bool {
		c := e.Data.(*engine.Client)
		assert.Equal(t, app.Client.ID, c.ID)

		done <- true
		return true
	})

	app.Start()
	<-done
}

func TestEventFiresOnDisconnection(t *testing.T) {
	app := NewApplicationTest("test-122")
	done := make(chan bool)

	app.GM.RegisterHandler("disconnected", func(e engine.Event) bool {
		done <- true
		return true
	})

	app.Start()
	app.Disconnect()
	<-done
}

func TestDirectEvent(t *testing.T) {
	app := NewApplicationTest("test-123")
	done := make(chan bool)
	app.GM.AddComponents(map[string]engine.ComponentInterface{
		"test": &TestEvents{},
	})

	go func() {
		for {
			select {
			case e, _ := <-app.Connection.In:
				if e.Name == "test.direct" {
					done <- true
				}
			}
		}
	}()

	app.Start()
	app.GM.FireEvent(engine.NewDirectEvent("test.direct", "", "test-123"))
	<-done
}
