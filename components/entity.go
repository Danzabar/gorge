package components

import (
    "github.com/Danzabar/gorge/engine"
    "github.com/jinzhu/gorm"
)

type (
    // EntityComponent provides some useful tools when dealing
    // with automatically updating entities
    EntityComponent struct {
        engine.Component
    }

    EntityEvent struct {
        Value     interface{} `json:"value"`
        Type      string      `json:"type"`
        Operation string      `json:"operation"`
    }
)

func (e *EntityComponent) Register() {
    e.Event("entity:created", "fired when a new entity is created", []string{engine.INTERNAL_CHAN, engine.SERVER_CHAN})

    // Register the gorm callbacks
    e.callbacks()
}

// OnCreate fires a new created event when an entity is created
func (e *EntityComponent) OnCreate(scope *gorm.Scope) {
    e.Fire("entity:created", e.createEventFromScope(scope))
}

// Creates an entity event from given scope and type
func (e *EntityComponent) createEventFromScope(scope *gorm.Scope) *EntityEvent {
    return &EntityEvent{
        Value: scope.Value,
        Type:  scope.IndirectValue().Type().Name(),
    }
}

// Registers the callbacks for gorm to allow entities to auto-update
// accross the websocket connection
func (e *EntityComponent) callbacks() {
    db := e.DB()

    // Add scoped callbacks
    db.Callback().Create().After("gorm:create").Register("gorge:create", e.OnCreate)
}
