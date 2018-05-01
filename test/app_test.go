package test

import (
	"github.com/Danzabar/gorge/engine"
)

type (
	// ApplicationTest is a wrapper that connects a client
	// and the game manager
	ApplicationTest struct {
		GM     *engine.GameManager
		Client *engine.Client
	}

	// TestConnection is used to connect to the server
	// the out channel can be used to simulate client side events
	// and the in slice allows you to see what events have been sent
	// from server to client
	TestConnection struct {
		In  []engine.Event
		Out chan engine.Event
	}
)

// Writer is required by ConnectionInterface, this stores the messages in
// the In slice
func (t *TestConnection) Writer(c *engine.Client, s *engine.Server) {
writerloop:
	for {
		select {
		case event, ok := <-c.Send:
			if !ok {
				break writerloop
			}

			// Otherwise direct it straight to the IN channel
			t.In = append(t.In, event)

			break
		}
	}
}

// Reader is required by ConnectionInterface, this simulates sending a
// message from the client to the server
func (t *TestConnection) Reader(c *engine.Client, s *engine.Server) {
	for {
		select {
		case e := <-t.Out:
			e.ClientId = c.Id
			e.Origin = engine.ORIG_CLIENT
			s.GM.FireEvent(e)
			break
		}
	}
}

// NewApplicationTest creates a new test application
func NewApplicationTest(c string) *ApplicationTest {
	return &ApplicationTest{
		GM:     engine.NewGame(),
		Client: engine.NewClient(&TestConnection{}, c),
	}
}

// StartNewAppTest creates a blank application and starts it straight away
func StartNewAppTest(c string) *ApplicationTest {
	app := NewApplicationTest(c)

	app.Start()
	return app
}

// Start runs the game manager and registers the client
func (a *ApplicationTest) Start() {
	// Run the app
	a.GM.Run()

	// Register the client
	a.GM.Server.Register <- a.Client
}
