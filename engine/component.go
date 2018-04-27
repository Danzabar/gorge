package engine

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type (

	// Interface for components
	ComponentInterface interface {
		SetGM(*GameManager)
		// Method to register events for this component.
		//
		// This can be used to register events, perform actions, start
		// separate processes, the method is only executed once, just
		// before the server is started.
		Register()
		// Setup is the first method called, called when adding a component
		// it is responsible for setting things up way before the call to start
		// the game
		//
		// Since set up is the first method called, we use this to load
		// the component pointer, this allows us to use the helper methods
		// on the component without harming the autoload process.
		Setup()
	}

	// Component is a base for components that accepts
	// the game manager and provides an interface for
	// easier usage.
	Component struct {
		GM *GameManager
	}
)

// SetGM sets a reference to the GameManager pointer
func (c *Component) SetGM(GM *GameManager) {
	c.GM = GM
}

// Register default register method for components that do not require it
func (c *Component) Register() {}

// Setup default setup method for components that do not require it
func (c *Component) Setup() {}

// Log returns the logrus instance from the application
func (c *Component) Log() *logrus.Logger {
	return c.GM.Log
}

func (c *Component) DB() *mgo.Database {
	return c.GM.DB.Instance()
}

func (c *Component) Save(n string, i interface{}) {
	c.GM.DB.Save(n, i)
}

// GetConfigAs deserializes the raw output of a config into
// the given interface
func (c *Component) GetConfigAs(n string, i interface{}) error {
	return c.GM.Config.ConvertYaml(n, i)
}

// GetClient finds a client by its identifier
func (c *Component) GetClient(n string) (*Client, error) {
	return c.GM.Server.Find(n)
}

// ClientValid checks whether a client id is valid
func (c *Component) ClientValid(n string) bool {
	if _, err := c.GM.Server.Find(n); err != nil {
		return false
	}

	return true
}

// GetChannel proxy method to fetch a channel from the server
func (c *Component) GetChannel(n string) (ChannelInterface, error) {
	return c.GM.Server.FindChannel(n)
}

// RemoveTrait proxy method
func (c *Component) RemoveTrait(n string, cl *Client) {
	c.GM.RemoveTrait(n, cl)
}

// PutTrait proxy method
func (c *Component) PutTrait(n string, t TraitInterface, cl *Client) {
	c.GM.PutTrait(n, t, cl)
}

// Event is an easier to use proxy method to register an event
func (c *Component) Event(n string, s string, strict bool, ch []string) {
	c.GM.Event(EventDefinition{n, s, strict, true, ch})
}

// Register proxy method to register a new event handler
func (c *Component) Handler(n string, h EventHandler) {
	c.GM.RegisterHandler(n, h)
}

// FireEvent is a proxy method for the managers Fire event
func (c *Component) Fire(n string, d interface{}) {
	c.GM.FireEvent(NewEvent(n, d))
}

// FireTo is a proxy method to fire a new direct event
func (c *Component) FireTo(n string, cl string, d interface{}) {
	c.GM.FireEvent(NewDirectEvent(n, d, cl))
}

// ConnectTo is a proxy method for the servers connect to channel method
func (c *Component) ConnectTo(n string, client *Client) {
	c.GM.Server.ConnectTo(n, client)
}

// Channel creates a new channel on the server
func (c *Component) Channel(n string, ch ChannelInterface) {
	c.GM.Server.NewChannels(map[string]ChannelInterface{n: ch})
}

// Channels creates channels using the given map
func (c *Component) Channels(ch map[string]ChannelInterface) {
	c.GM.Server.NewChannels(ch)
}
