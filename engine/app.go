package engine

import (
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type (

	// Gamemanager handles components and subscriptions
	GameManager struct {
		Config      *ConfigManager
		Settings    *GorgeSettings
		DB          *Mongo
		Components  *sync.Map
		Instances   *sync.Map
		Subscribers *sync.Map
		Events      *sync.Map
		Server      *Server
		Log         *logrus.Logger
		Registry    *Registry
	}
)

// NewGame creates a new instance of the game manager
func NewGame() *GameManager {

	GM := &GameManager{
		Components:  new(sync.Map),
		Instances:   new(sync.Map),
		Subscribers: new(sync.Map),
		Events:      new(sync.Map),
		Log:         NewLog(),
		Registry:    NewRegistry(),
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

	// Auto load from config
	GM.Autoload()

	// Load Mongo
	GM.CreateMongo()

	// Set up Components
	defer GM.RegisterComponents()

	GM.Log.Info("Game has started...")

	// Start the servers listen routine, so we can connect
	// to it
	GM.Server.Listen()
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

func (GM *GameManager) Autoload() {
	// Load the dynamic traits for the confi
	for _, v := range GM.Settings.Traits {
		// Check if they are in the registry
		i, err := GM.Registry.GetStruct(v.Name)

		if err != nil {
			GM.Log.Error(err)
			continue
		}

		// If it exists, register it
		GM.RegisterTrait(map[string]TraitInterface{v.Name: i.(TraitInterface)})
	}
}

// BindTrait binds a registered instance to a client
func (GM *GameManager) BindTrait(n string, c *Client) {
	// Does the given instance exist?
	i, ok := GM.Instances.Load(n)

	if !ok {
		GM.Log.Errorf("Unable to load instance with given name %s", n)
		return
	}

	// Create a new instance
	inst := NewInstance(GM)
	in := i.(TraitInterface)

	// Bind to connection
	c.BindTrait(n, in, inst)
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

		GM.Instances.Store(key, value)
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

	go GM.Server.SendToChannels(e, definition)
}
