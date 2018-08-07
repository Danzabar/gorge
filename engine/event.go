package engine

import (
	_ "io/ioutil"
	"time"

	"github.com/teris-io/shortid"
	_ "github.com/xeipuuv/gojsonschema"
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
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Channels    []string `json:"channels"`
	}

	// EventHandler is used to process events
	EventHandler func(e Event) bool

	// EventValidator is the contact for a validator
	EventValidator interface {
		// Validate is used to validate the message against
		// a schema
		Validate(schema string, subject interface{}) error
	}
)

// Validate validates the integrity of a message against a schema
/*func (e EventDefinition) Validate(p interface{}) (bool, error) {
	rs, err := ioutil.ReadFile(e.Schema)

	if err != nil {
		return false, err
	}

	s := gojsonschema.NewStringLoader(string(rs))
	d := gojsonschema.NewGoLoader(p)

	result, err := gojsonschema.Validate(s, d)

	return result.Valid(), err
}*/

// NewEvent creates a new event
func NewEvent(n string, d interface{}) Event {
	id, _ := shortid.Generate()

	return Event{
		ID:        id,
		Name:      n,
		Data:      d,
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
