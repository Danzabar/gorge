package engine

import (
	"time"
)

type (
	Entity struct {
		Id        string    `bson:"_id,omitempty" json:"id,omitempty"`
		ClientId  string    `bson:"clientId" json:"clientId"`
		CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
		UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
	}

	EntityInterface interface {
		SetClientId(string)
		OnCreate(string, *GameManager)
		OnUpdate(*GameManager)
		OnDelete(*GameManager)
	}
)

func (e *Entity) OnCreate(id string, GM *GameManager) {
	// Set the ID first
	e.Id = id

	// Update the timestamps
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
}

func (e *Entity) OnUpdate(GM *GameManager) {
	// Update the timestamps
	e.UpdatedAt = time.Now()
}

func (e *Entity) OnDelete(GM *GameManager) {}

// SetClientId is default handler to set client id on an Entity
func (e *Entity) SetClientId(in string) {
	e.ClientId = in
}
