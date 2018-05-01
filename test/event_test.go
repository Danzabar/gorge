package test

import (
	"testing"

	"github.com/Danzabar/gorge/engine"
)

func TestEventFiresOnConnection(t *testing.T) {
	app := NewApplicationTest("test-122")
	done := make(chan bool)

	app.GM.RegisterHandler("connected", func(e engine.Event) bool {
		done <- true
		return true
	})

	app.Start()
	<-done
}
