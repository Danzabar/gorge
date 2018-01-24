package engine

import (
    "time"

    "github.com/teris-io/shortid"
)

const (
    // Constant value for an internal origin
    ORIG_INTERNAL = "internal"

    // Constant value for an inbound origin
    ORIG_CLIENT = "client"
)

type (
    // Event contains a platform event details
    Event struct {
        ID         string          `json:"id"`
        Name       string          `json:"name"`
        Data       interface{}     `json:"data"`
        Broadcast  bool            `json:"broadcast"`
        Origin     string          `json:"origin"`
        Definition EventDefinition `json:"definition"`
        ClientId   string          `json:"clientId"`
        CreatedAt  time.Time       `json:"createdAt"`
    }

    // EventDefinition stores the definition of an event
    // used for documentation
    EventDefinition struct {
        Name        string   `json:"name"`
        Description string   `json:"description"`
        Channels    []string `json:"channels"`
    }

    // EventHandlers are used to process events
    EventHandler func(e Event) bool
)

func NewEvent(n string, d interface{}) Event {
    id, _ := shortid.Generate()

    return Event{
        ID:        id,
        Name:      n,
        Data:      d,
        Broadcast: false,
        Origin:    ORIG_INTERNAL,
        CreatedAt: time.Now(),
    }
}

func NewDirectEvent(n string, d interface{}, c string) Event {
    ev := NewEvent(n, d)
    ev.ClientId = c
    return ev
}
