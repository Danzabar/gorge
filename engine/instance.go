package engine

type (
	// Instance is a helper struct that contains all the relevent
	// information an instance might need
	Instance struct {
		*Component
		Client *Client
	}

	// InstanceInterface defines the expectations of an instance
	InstanceInterface interface {
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
