package engine

type (
	Instance struct {
		*Component
	}

	// InstanceInterface defines the expectations of an instance
	InstanceInterface interface {
		Register(*Instance)
		Connect(*Instance)
		Destroy(*Instance)
	}
)

func NewInstance(GM *GameManager) *Instance {
	i := &Instance{}
	i.SetGM(GM)
	return i
}

// Handler is an override to create a new event handler,
// this ensures that an instance can bind to direct event messages
func (i *Instance) Handler(n string, h EventHandler) {
	i.Component.GM.InstanceHandler(n, h)
}
