package engine

import (
	"io/ioutil"
	"time"

	"github.com/teris-io/shortid"
)

const (
	// InternalOrigin sconstant value for an internal origin
	InternalOrigin = "internal"

	// ClientOrigin constant value for an inbound origin
	ClientOrigin = "client"
)

type (
	// Event contains a platform event details
	Event struct {
		ID        string      `json:"id"`
		Name      string      `json:"name"`
		Data      interface{} `json:"data"`
		Broadcast bool        `json:"broadcast"`
		Origin    string      `json:"origin"`
		ClientID  string      `json:"clientId"`
		CreatedAt time.Time   `json:"createdAt"`
	}

	// EventDefinition stores the definition of an event
	// used for documentatio
	EventDefinition struct {
		Name        string
		Description string
		Channels    []string
		Validator   EventValidator
	}

	// EventValidator allows the attaching of a validator
	// and a schema to an event definition
	EventValidator struct {
		Handler Validation
		Schema  string
	}

	// EventHandler is used to process events
	EventHandler func(e Event) bool

	// Validation is a contract for a func that can handle
	// validation rules given a schema and data
	Validation func(schema string, subject interface{}) error
)

// Validate validates the integrity of a message against a schema
func (e EventDefinition) Validate(p interface{}) error {
	if e.Validator.Handler != nil {
		if err := e.Validator.Handler(e.Validator.Schema, p); err != nil {
			return err
		}
	}

	return nil
}

// NewEvent creates a new event
func NewEvent(name string, data interface{}) Event {
	id, _ := shortid.Generate()

	return Event{
		ID:        id,
		Name:      name,
		Data:      data,
		Broadcast: false,
		Origin:    InternalOrigin,
		CreatedAt: time.Now(),
	}
}

// NewDirectEvent creates a new direct event
func NewDirectEvent(n string, d interface{}, c string) Event {
	ev := NewEvent(n, d)
	ev.ClientID = c
	return ev
}

// NewEventValidator creates a new event validator given a file name
func NewEventValidator(file string, h Validation) EventValidator {
	rs, err := ioutil.ReadFile(file)

	if err != nil {
		panic(err)
	}

	return EventValidator{Handler: h, Schema: string(rs)}
}
