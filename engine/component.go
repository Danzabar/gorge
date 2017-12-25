package engine

type (

    // Interface for components
    ComponentInterface interface {
        // Method to set the reference of the GameManager
        SetGM(*GameManager)
        // Method to register events for this component
        Register()
    }

    // Component is a base for components that accepts
    // the game manager and provides an interface for
    // easier usage.
    //
    Component struct {
        GM *GameManager
    }
)

// SetGM sets a reference to the GameManager pointer
func (c *Component) SetGM(GM *GameManager) {
    c.GM = GM
}

// Event is an easier to use proxy method to register an event
func (c *Component) Event(n string, d string, bc bool, ws bool, in bool) {
    c.GM.Event(EventDefinition{n, d, bc, ws, in})
}

// Register proxy method to register a new event handler
func (c *Component) Register(n string, h EventHandler) {
    c.GM.RegisterHandler(n, h)
}

// FireEvent is a proxy method for the managers Fire event
func (c *Component) Fire(n string, d interface{}) {
    c.GM.FireEvent(NewEvent(n, d))
}
