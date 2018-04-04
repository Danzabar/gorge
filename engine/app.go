package engine

import (
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type (

	// Gamemanager handles components and subscriptions
	GameManager struct {
		Config      *ConfigManager
		Settings    *GorgeSettings
		DB          *Mongo
		Components  *sync.Map
		Subscribers *sync.Map
		Events      *sync.Map
		Server      *Server
		Log         *logrus.Logger
	}
)

// NewGame creates a new instance of the game manager
func NewGame() *GameManager {

	GM := &GameManager{
		Components:  new(sync.Map),
		Subscribers: new(sync.Map),
		Events:      new(sync.Map),
		Log:         NewLog(),
	}

	GM.Config = NewConfig(GM)
	GM.Server = NewServer(GM)

	return GM
}

// NewLog creates a new logrus logger
func NewLog() *logrus.Logger {
	Log := logrus.New()
	Log.Out = os.Stdout
	Log.Level = logrus.DebugLevel

	return Log
}

// Run for now isn't much, but it will be incharge of
// initialising components and running the game engine
func (GM *GameManager) Run() {
	// Load Standard Configuration
	GM.Config.LoadStandard()

	// Load custom configuration
	GM.Config.Load()

	// Load Mongo
	GM.CreateMongo()

	defer func() {
		// Set up Components
		GM.RegisterComponents()

		GM.Log.Info("Game has started...")
	}()

	// Start the servers listen routine, so we can connect
	// to it
	go GM.Server.Listen()
}

// CreateMongo attaches a new mongo wrapper to the game manager
func (GM *GameManager) CreateMongo() {
	GM.DB = NewMongo(GM)
}

// Connect adds a new client with the given connection and
// identifier
func (GM *GameManager) Connect(ws *websocket.Conn, id string) {
	// Create the client
	c := NewClient(&WebsocketConnection{Conn: ws}, id)

	// Register it on the server
	GM.Server.Register <- c
}

func (GM *GameManager) DBInstance() *mgo.Database {
	return GM.DB.Instance().DB(GM.DB.Settings.Database)
}

// PutTrait binds an existing trait to a client
func (GM *GameManager) PutTrait(n string, t TraitInterface, c *Client) {
	inst := NewInstance(GM)
	c.BindTrait(n, t, inst)
}

// RemoveTrait removes the trait instance from the client
func (GM *GameManager) RemoveTrait(n string, c *Client) {
	inst := NewInstance(GM)
	c.RemoveTrait(n, inst)
}

// RegisterHandler registers a new event handler
func (GM *GameManager) RegisterHandler(n string, h EventHandler) {
	var handlers []EventHandler

	reg, ok := GM.Subscribers.Load(n)

	if ok {
		handlers = reg.([]EventHandler)
	}

	// append the new handler
	handlers = append(handlers, h)
	GM.Subscribers.Store(n, handlers)
}

// Event registers a new event definition, all events
// need to be registered before being fired
func (GM *GameManager) Event(e EventDefinition) {
	// Register the event
	GM.Events.Store(e.Name, e)
}

// AddComponents adds a map of components to the store
func (GM *GameManager) AddComponents(components map[string]ComponentInterface) {
	for key, value := range components {
		GM.Log.Infof("Loading component %s", key)
		value.SetGM(GM)
		value.Setup()

		// Add to the store
		GM.Components.Store(key, value)
	}
}

// RegisterInstance registers a new trait, this just stores a point to it
// which when connected to a client is copied
func (GM *GameManager) RegisterTrait(instances map[string]TraitInterface) {
	for key, value := range instances {
		GM.Log.Infof("Registering trait %s", key)
		value.Register(NewInstance(GM))
	}
}

// RegisterComponents calls the register method on
// components in the store
func (GM *GameManager) RegisterComponents() {
	GM.Components.Range(func(k, v interface{}) bool {
		component := v.(ComponentInterface)
		component.Register()
		GM.Log.Infof("Registered component %s", k.(string))
		return true
	})

	GM.Log.Info("Finished registering components...")
}

// FireEvent fires the event using the rules registered in the
// associative definition
func (GM *GameManager) FireEvent(e Event) {
	def, ok := GM.Events.Load(e.Name)

	if !ok {
		GM.Log.Errorf("Unable to locate a triggered event %s", e.Name)
		return
	}

	definition := def.(EventDefinition)
	e.Definition = definition

	// Does it have a schema and is it strict schema?
	if definition.StrictSchema && definition.Schema != "" {
		if ok, err := definition.Validate(e.Data); !ok {
			// At this point we cannot send to channels
			GM.Log.Error("Unable to send message as it does not adhere to schema")
			GM.Log.Error(err)
			return
		}
	}

	go GM.Server.SendToChannels(e, definition)
}
