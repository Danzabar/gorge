package engine

import (
	"time"
)

type (
	// Entity represents a savable object, this is the base
	// requirements of an entity, it also provides the needed
	// funcs to fall under the EntityInterface. See streaming.
	Entity struct {
		ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
		ClientID  string    `bson:"clientId" json:"clientId"`
		CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
		UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	}

	// EntityInterface a contract that allows an entity
	// the option of using the easier save and streaming
	// methods, provides hooks for create,save and delete.
	//
	// It can also be used to dynamically set the client the
	// entity belongs to.
	EntityInterface interface {
		SetClientID(string)
		OnCreate(string, *GameManager)
		OnUpdate(*GameManager)
		OnDelete(*GameManager)
	}
)

// OnCreate fired when the entity is first created,
// this sets the id and createdAt/updatedAt properties.
func (e *Entity) OnCreate(id string, GM *GameManager) {
	// Set the ID first
	e.ID = id

	// Update the timestamps
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
}

// OnUpdate fired whenever the entity is updated,
// this sets the updatedAt to the correct timestamp
func (e *Entity) OnUpdate(GM *GameManager) {
	// Update the timestamps
	e.UpdatedAt = time.Now()
}

// OnDelete - here merely to satisfy EntityInterface
func (e *Entity) OnDelete(GM *GameManager) {}

// SetClientID is default handler to set client id on an Entity
func (e *Entity) SetClientID(in string) {
	e.ClientID = in
}
