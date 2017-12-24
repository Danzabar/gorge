package engine

import (
    "sync"

    "github.com/sirupsen/logrus"
)

type (

    // Gamemanager handles components and subscriptions
    GameManager struct {
        Components  *sync.Map
        Subscribers *sync.Map
        Events      *sync.Map
        Log         *logrus.Logger
    }
)

// NewGame creates a new instance of the game manager
func NewGame() *GameManager {
    return &GameManager{
        Components:  new(sync.Map),
        Subscribers: new(sync.Map),
        Events:      new(sync.Map),
        Log:         NewLog(),
    }
}

// NewLog creates a new logrus logger
func NewLog() *logrus.Logger {
    Log = logrus.New()
    Log.Out = os.Stdout
    Log.Level = logrus.DebugLevel

    return Log
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

}
