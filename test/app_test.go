package test

import (
	"github.com/Danzabar/gorge/engine"
	"testing"
)

type (
	// ApplicationTest is a wrapper that connects a client
	// and the game manager
	ApplicationTest struct {
		GM         *engine.GameManager
		Client     *engine.Client
		Connection *TestConnection
	}

	// TestConnection is used to connect to the server
	// the out channel can be used to simulate client side events
	// and the in slice allows you to see what events have been sent
	// from server to client
	TestConnection struct {
		In  chan engine.Event
		Out chan engine.Event
	}

	TestEntity struct {
		engine.Entity
		Name string `json:"string"`
		Foo  string `json:"string"`
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
			t.In <- event

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
			e.ClientID = c.ID
			e.Origin = engine.ClientOrigin
			s.GM.FireEvent(e)
			break
		}
	}
}

func NewTestConnection() *TestConnection {
	return &TestConnection{
		In:  make(chan engine.Event),
		Out: make(chan engine.Event),
	}
}

// NewApplicationTest creates a new test application
func NewApplicationTest(c string) *ApplicationTest {
	conn := NewTestConnection()

	return &ApplicationTest{
		GM:         engine.NewGame(),
		Client:     engine.NewClient(conn, c),
		Connection: conn,
	}
}

// StartNewAppTest creates a blank application and starts it straight away
func StartNewAppTest(c string) *ApplicationTest {
	app := NewApplicationTest(c)

	app.Start()
	return app
}

// Disconnect fires the disconnect event to the server
func (a *ApplicationTest) Disconnect() {
	a.GM.Server.Disconnect(a.Client)
}

// Start runs the game manager and registers the client
func (a *ApplicationTest) Start() {
	// Run the app
	a.GM.Run()

	// Register the client
	a.GM.Server.Register <- a.Client
}

func BenchmarkLoadApplication(b *testing.B) {
	for i := 0; i < b.N; i++ {
		app := NewApplicationTest("tester")
		app.GM.Run()
	}
}

func BenchmarkLoadApplicationWithComponents(b *testing.B) {
	for i := 0; i < b.N; i++ {
		app := NewApplicationTest("tester")
		app.GM.AddComponents(map[string]engine.ComponentInterface{
			"test1": &TestEvents{},
			"test2": &TestEvents{},
			"test3": &TestEvents{},
		})
		app.GM.Run()
	}
}

func BenchmarkEventSend(b *testing.B) {
	app := NewApplicationTest("tester")
	app.GM.AddComponents(map[string]engine.ComponentInterface{
		"events": &TestEvents{},
	})

	app.Start()

	for i := 0; i < b.N; i++ {
		app.GM.FireEvent(engine.NewDirectEvent("test.direct", "", "tester"))
	}
}

func BenchmarkStreamEventOutboundSend(b *testing.B) {
	app := NewApplicationTest("tester")
	app.GM.StreamManager.New("test", "test", &TestEntity{}, false)

	app.Start()
	app.GM.DB.Settings = engine.MongoSettings{Host: "localhost", Database: "test", AutoConnect: true}
	app.GM.DB.Connect()

	for i := 0; i < b.N; i++ {
		t := &TestEntity{Name: "test", Foo: "bar"}
		app.GM.DB.Save("test", t)
	}
}

func BenchmarkStreamEventFullCycle(b *testing.B) {
	app := NewApplicationTest("tester")
	app.GM.StreamManager.New("test", "test", &TestEntity{}, false)

	app.Start()
	app.GM.DB.Settings = engine.MongoSettings{Host: "localhost", Database: "test", AutoConnect: true}
	app.GM.DB.Connect()

	for i := 0; i < b.N; i++ {
		done := make(chan bool)

		// Server responds with a stream event
		go func() {
			for {
				select {
				case e, _ := <-app.Connection.In:
					app.GM.Log.Error(e.Name)
					if e.Name == "stream.updated" {
						done <- true
					}
				}
			}
		}()

		// User sends in stream event
		e := engine.Event{
			Name: "stream.save",
			Data: engine.StreamSchema{
				Stream:   "test",
				Data:     &TestEntity{Name: "Dave"},
				ClientID: "tester"},
			ClientID: "tester",
		}

		app.Connection.Out <- e

		// The test isn't finished until we get the event
		<-done
	}
}
