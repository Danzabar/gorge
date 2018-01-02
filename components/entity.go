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

    // EntityEvent represents the entity value and type
    // of time of change
    EntityEvent struct {
        Value interface{} `json:"value"`
        Type  string      `json:"type"`
    }
)

func (e *EntityComponent) Register() {
    // Register events
    e.Event("entity:created", "fired when a new entity is created", []string{engine.INTERNAL_CHAN, engine.SERVER_CHAN})
    e.Event("entity:updated", "fired when an entity is updated", []string{engine.INTERNAL_CHAN, engine.SERVER_CHAN})
    e.Event("entity:deleted", "fired when an entity is deleted", []string{engine.INTERNAL_CHAN, engine.SERVER_CHAN})

    // Register the gorm callbacks
    e.callbacks()
}

// OnCreate fires a new created event when an entity is created
func (e *EntityComponent) OnCreate(scope *gorm.Scope) {
    e.Fire("entity:created", e.createEventFromScope(scope))
}

// OnUpdate fires a new updated event when an entity is updated
func (e *EntityComponent) OnUpdate(scope *gorm.Scope) {
    e.Fire("entity:updated", e.createEventFromScope(scope))
}

// OnDelete fires a new deleted event when an entity is deleted
func (e *EntityComponent) OnDelete(scope *gorm.Scope) {
    e.Fire("entity:deleted", e.createEventFromScope(scope))
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
    db.Callback().Delete().After("gorm:delete").Register("gorge:delete", e.OnDelete)
    db.Callback().Update().After("gorm:update").Register("gorge:update", e.OnUpdate)
}
