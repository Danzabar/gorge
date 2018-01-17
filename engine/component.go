package engine

import (
    "github.com/jinzhu/gorm"
    "github.com/sirupsen/logrus"
)

type (

    // Interface for components
    ComponentInterface interface {
        // Method to set the reference of the GameManager
        SetGM(*GameManager)
        // Method to register events for this component
        Register()
        // Method to allow the component to run along side the application
        // the reason this is seperated is to allow for more configuration
        // checks before running in the future
        Run()
        // Setup is the first method called, called when adding a component
        // it is responsible for setting things up way before the call to start
        // the game
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

// Run default run method for components that do not require it
func (c *Component) Run() {}

// Setup default setup method for components that do not require it
func (c *Component) Setup() {}

// Log returns the logrus instance from the application
func (c *Component) Log() *logrus.Logger {
    return c.GM.Log
}

// DB fetches the database instance from the application
func (c *Component) DB() *gorm.DB {
    return c.GM.DB
}

// Migrate registers a migration entity
func (c *Component) Migrate(i interface{}) {
    c.GM.AddMigration(i)
}

// Event is an easier to use proxy method to register an event
func (c *Component) Event(n string, d string, ch []string) {
    c.GM.Event(EventDefinition{n, d, ch})
}

// Register proxy method to register a new event handler
func (c *Component) Handler(n string, h EventHandler) {
    c.GM.RegisterHandler(n, h)
}

// FireEvent is a proxy method for the managers Fire event
func (c *Component) Fire(n string, d interface{}) {
    c.GM.FireEvent(NewEvent(n, d))
}

// Channel creates a new channel on the server
func (c *Component) Channel(n string, ch ChannelInterface) {
    c.GM.Server.NewChannels(map[string]ChannelInterface{n: ch})
}

func (c *Component) Channels(ch map[string]ChannelInterface) {
    c.GM.Server.NewChannels(ch)
}
