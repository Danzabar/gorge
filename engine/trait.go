package engine

type (
	// Trait is a helper struct that contains all the relevent
	// information an instance might need
	Trait struct {
		Component
		Client *Client
	}

	// TraitInterface defines the expectations of an instance
	TraitInterface interface {
		// SetGM sets the GameManager instance, allowing extra functionality
		SetGM(*GameManager)
		// SetClient sets the relevent client to the trait
		SetClient(*Client)
		// Register is responsible for setting up an instance before
		// its connected
		Register()
		// Connect is called when the instance is bound to the client
		Connect()
		// Destroy is called when the instance is destroyed
		Destroy()
	}
)

// Register default
func (t *Trait) Register() {}

// Connect default
func (t *Trait) Connect() {}

// Destroy default
func (t *Trait) Destroy() {}

// SetClient sets the connected client
func (t *Trait) SetClient(c *Client) {
	t.Client = c
}

// Handler is an override to create a new event handler,
// this ensures that an instance can bind to direct event messages
func (t *Trait) Handler(n string, h EventHandler) {
	t.Client.RegisterHandler(n, h)
}
