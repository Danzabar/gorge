package engine

import (
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
)

type (

	// Gamemanager handles components and subscriptions
	GameManager struct {
		Config      *ConfigManager
		Settings    *GorgeSettings
		Components  *sync.Map
		Instances   *sync.Map
		Subscribers *sync.Map
		Events      *sync.Map
		Server      *Server
		Log         *logrus.Logger
		DB          *gorm.DB
		Migrations  []interface{}
	}
)

// NewGame creates a new instance of the game manager
func NewGame(dbDriver, dbCreds string) *GameManager {
	db, err := gorm.Open(dbDriver, dbCreds)

	if err != nil {
		panic(err)
	}

	GM := &GameManager{
		Components:  new(sync.Map),
		Instances:   new(sync.Map),
		Subscribers: new(sync.Map),
		Events:      new(sync.Map),
		Log:         NewLog(),
		DB:          db,
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

	// Run Migrations
	GM.Migrate()
	defer GM.RegisterComponents()

	// Run Components
	defer GM.RunComponents()

	GM.Log.Debug("Game has started...")

	// Start the servers listen routine, so we can connect
	// to it
	go GM.Server.Listen()
}

// Connect adds a new client with the given connection and
// identifier
func (GM *GameManager) Connect(ws *websocket.Conn, id string) {
	// Create the client
	c := NewClient(&WebsocketConnection{Conn: ws}, id)

	// Register it on the server
	GM.Server.Register <- c
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
		GM.Log.Debugf("Loading component %s", key)
		value.SetGM(GM)
		value.Setup()

		// Add to the store
		GM.Components.Store(key, value)
	}
}

// RegisterInstance registers a new instance, this just stores a point to it
// which when connected to a client is copied
func (GM *GameManager) RegisterTrait(instances map[string]TraitInterface) {
	for key, value := range instances {
		GM.Log.Debugf("Registering instance %s", key)
		value.Register(NewInstance(GM))

		GM.Instances.Store(key, value)
	}
}

// AddMigration adds an entity to the list of entities
// that will be migrated when the game runs
func (GM *GameManager) AddMigration(i interface{}) {
	GM.Migrations = append(GM.Migrations, i)
}

// RegisterComponents calls the register method on
// components in the store
func (GM *GameManager) RegisterComponents() {
	GM.Components.Range(func(k, v interface{}) bool {
		component := v.(ComponentInterface)
		component.Register()
		GM.Log.Debugf("Registered component %s", k.(string))
		return true
	})

	GM.Log.Debugf("Finished registering components...")
}

// RunComponents calls the run method on all registered
// components
func (GM *GameManager) RunComponents() {
	GM.Components.Range(func(k, v interface{}) bool {
		component := v.(ComponentInterface)
		go component.Run()
		GM.Log.Debugf("Starting component %s", k.(string))
		return true
	})

	GM.Log.Debug("All components have been started...")
}

// Migrate runs an automigration against the migration list
func (GM *GameManager) Migrate() {
	GM.Log.Debugf("Calling migrations...")
	for _, v := range GM.Migrations {
		GM.DB.AutoMigrate(v)
	}
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
