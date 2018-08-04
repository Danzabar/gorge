package engine

import (
	"flag"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

const (
	// TestEnv defines when the app is running in test mode
	TestEnv = "test"
	// ProdEnv defines when the app is running in production mode
	ProdEnv = "production"
	// DevEnv defines when the app is running in development mode
	DevEnv = "development"
)

type (

	// GameManager handles components and subscriptions
	GameManager struct {
		Environment   string
		Config        *ConfigManager
		Settings      *GorgeSettings
		DB            *Mongo
		Components    *sync.Map
		Subscribers   *sync.Map
		Events        *sync.Map
		Server        *Server
		StreamManager *StreamManager
		Log           *logrus.Logger
	}
)

// NewGame creates a new instance of the game manager
func NewGame() *GameManager {
	GM := &GameManager{
		Components:  new(sync.Map),
		Subscribers: new(sync.Map),
		Events:      new(sync.Map),
		Log:         NewLog(),
		Environment: environment(),
	}

	// Should be extracted to a seperate method at some point
	if GM.Environment == TestEnv {
		GM.Log.Level = logrus.WarnLevel
	}

	GM.Config = NewConfig(GM)
	GM.Server = NewServer(GM)
	GM.StreamManager = NewStreamManager(GM)

	return GM
}

// Defines which environment we are in by checking
// first if we are in test, and secondly for env vars
// this will default to development
func environment() string {
	if flag.Lookup("test.v") != nil {
		godotenv.Load("../.env-test")
		return TestEnv
	}

	// TODO: Add production env detection

	return DevEnv
}

// NewLog creates a new logrus logger
func NewLog() *logrus.Logger {
	Log := logrus.New()
	Log.Out = os.Stdout
	Log.Level = logrus.DebugLevel

	return Log
}

// Decode ensures an event is decoded properly
func Decode(in interface{}, out interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           out,
		WeaklyTypedInput: false,
		ErrorUnused:      false,
		DecodeHook:       mapstructure.StringToTimeHookFunc(time.RFC3339),
	}

	decoder, err := mapstructure.NewDecoder(config)

	if err != nil {
		panic(err)
	}

	return decoder.Decode(in)
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
		// Register Stream events
		GM.StreamManager.Register()

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

// PutTrait binds an existing trait to a client
func (GM *GameManager) PutTrait(n string, t TraitInterface, c *Client) {
	t.SetGM(GM)
	c.BindTrait(n, t)
}

// RemoveTrait removes the trait instance from the client
func (GM *GameManager) RemoveTrait(n string, c *Client) {
	c.RemoveTrait(n)
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

// RegisterTrait registers a new trait, this just stores a point to it
// which when connected to a client is copied
func (GM *GameManager) RegisterTrait(instances map[string]TraitInterface) {
	for key, value := range instances {
		GM.Log.Infof("Registering trait %s", key)
		value.SetGM(GM)
		value.Register()
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

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			GM.Log.Error(r)
		}
	}()

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
