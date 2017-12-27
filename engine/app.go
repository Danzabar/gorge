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
    var wg sync.WaitGroup

    wg.Add(1)

    // Register Components
    go GM.RegisterComponents()

    wg.Wait()
    GM.Log.Debug("Game has started...")
}

// Connect adds a new client with the given connection and
// identifier
func (GM *GameManager) Connect(ws *websocket.Conn, id string) {
    // Create the client
    c := &Client{
        Id:   id,
        Conn: ws,
        Send: make(chan Event),
    }

    // Register it on the server
    GM.Server.Register <- c
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

        // Add to the store
        GM.Components.Store(key, value)
    }
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

    go GM.Server.SendToChannels(e, definition)
}
