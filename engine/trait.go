package engine

import (
	"gopkg.in/mgo.v2/bson"
)

type (
	// Instance is a helper struct that contains all the relevent
	// information an instance might need
	Instance struct {
		Component
		Client *Client
	}

	// Trait represents the base contract for a trait
	// it can be used solely as a trait although it will not
	// provide any useful functionality.
	//
	// The trait struct also provides a useful base for the
	// mongo integration, if autoSave is enabled traits will
	// be periodically saved by the server, if autosave is not
	// enabled, traits can still be saved manually.
	Trait struct {
		Id        bson.ObjectId `bson:"_id" json:"-"`
		ClientId  string        `bson:"clientId" json:"clientId"`
		CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
		UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
	}

	// TraitInterface defines the expectations of an instance
	TraitInterface interface {
		// Register is responsible for setting up an instance before
		// its connected
		Register(*Instance)
		// Connect is called when the instance is bound to the client
		Connect(*Instance)
		// Destroy is called when the instance is destroyed
		Destroy(*Instance)
	}
)

// NewInstance creates a new instance
func NewInstance(GM *GameManager) *Instance {
	i := &Instance{}
	i.SetGM(GM)
	return i
}

// SetClient sets the connected client
func (i *Instance) SetClient(c *Client) {
	i.Client = c
}

// Handler is an override to create a new event handler,
// this ensures that an instance can bind to direct event messages
func (i *Instance) Handler(n string, h EventHandler) {
	i.Client.RegisterHandler(n, h)
}
