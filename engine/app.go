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

// FireEvent fires the event using the rules registered in the
// associative definition
func (GM *GameManager) FireEvent(e Event) {
    def, ok := GM.Events.Load(e.Name)
    definition := def.(EventDefinition)

    if !ok {
        GM.Log.Errorf("Unable to locate a triggered event %s", e.Name)
        return
    }

    // Check the details of the definition
    if definition.Internal {
        go GM.sendInternalEvent(e)
    }

    if definition.Websocket {
        go GM.sendWebsocketEvent(e, definition.Broadcast)
    }
}

// Sends an internal event, uses the subscribers list
func (GM *GameManager) sendInternalEvent(e Event) {
    subs, ok := GM.Subscribers.Load(e.Name)

    if !ok {
        GM.Log.Warningf("Event called with no active subscribers: %s", e.Name)
        return
    }

    subscribers := subs.([]EventHandler)

    // Fire all the things
    for _, sub := range subscribers {
        sub(e)
    }
}

// Either direct or broadcast to a client/s
func (GM *GameManager) sendWebsocketEvent(e Event, b bool) {
    // If its a broadcast, we don't need to find a client
    if b {
        GM.Server.Broadcast(e)
        return
    }

    if !b && e.ClientId == "" {
        GM.Log.Errorf("Direct event sent with no client id: %+v", e)
        return
    }

    // However if its not a broadcast, we need to have a client
    client, err := GM.Server.Find(e.ClientId)

    if err != nil {
        GM.Log.Errorf("Direct event sent with unknown client: %s", e.ClientId)
        return
    }

    // Otherwise we can send
    client.Send <- e
}
