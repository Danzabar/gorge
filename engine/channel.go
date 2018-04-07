package engine

import (
	"sync"
)

const (

	// Const value for the name of the internal channel
	INTERNAL_CHAN = "internal"

	// Const value for the name of the direct channel
	DIRECT_CHAN = "direct"

	// Const value for the name of the server channel
	SERVER_CHAN = "server"
)

type (

	// ChannelInterface is a blueprint for channels
	ChannelInterface interface {
		// Open is used to prepare the struct or register
		// any services the channel needs to operate
		Open()

		// Send is whats called when a new event is triggered
		// using the channel
		Send(Event, EventDefinition)

		// Close is called when the channel is closed
		Close()

		// Connect is called when a new client connects to the channel
		Connect(*Client)

		// Disconnect is called when a client exits the channel
		Disconnect(*Client)

		// SetGM sets the game manager instance
		SetGM(*GameManager)
	}

	// Channel is a base channel struct, helps creating a standard
	// channel
	Channel struct {
		GM      *GameManager
		Clients *sync.Map
	}

	StreamChannel struct {
		Channel
	}

	// InternalChannel works with component-component events
	InternalChannel struct {
		Channel
	}

	// ServerChannel is a channel that broadcasts by default
	ServerChannel struct {
		Channel
	}

	DirectChannel struct {
		Channel
	}
)

// SendToClients is a method that takes a list of clients to proxy an event to
func SendToClients(GM *GameManager, clients *sync.Map, e Event) {
	// If the message is a broadcast send it to everyone
	if e.Broadcast {
		clients.Range(func(k, v interface{}) bool {
			client := v.(*Client)

			client.Send <- e
			return true
		})

		return
	}

	if e.ClientId == "" {
		GM.Log.Errorf("Direct event sent with no client id: %+v", e)
		return
	}

	cl, ok := clients.Load(e.ClientId)

	if !ok {
		GM.Log.Errorf("Unable to find client from given id: %s", e.ClientId)
		return
	}

	client := cl.(*Client)
	client.Send <- e
}

// SendToTraits sends messages to traits
func SendToTraits(client *Client, e Event) {
	subs, ok := client.Subscribers.Load(e.Name)

	if ok {
		subscribers := subs.([]EventHandler)

		for _, sub := range subscribers {
			sub(e)
		}
	}
}

func (ch *Channel) SetGM(GM *GameManager) {
	ch.GM = GM
}

// Open - On the base channel object, this isn't really needed
func (ch *Channel) Open() {
	ch.Clients = new(sync.Map)
}

// Send sends an event to clients on the channel
func (ch *Channel) Send(e Event, d EventDefinition) {
	SendToClients(ch.GM, ch.Clients, e)
}

// Close - On the base channel object, this isn't really needed
func (ch *Channel) Close() {}

// Connect adds a new client to the list
func (ch *Channel) Connect(c *Client) {
	ch.Clients.Store(c.Id, c)
}

// Discconect removes the client
func (ch *Channel) Disconnect(c *Client) {
	ch.Clients.Delete(c.Id)
}

func (ch *StreamChannel) Send(e Event, d EventDefinition) {
	var schema StreamSchema
	Decode(e.Data, &schema)

	if schema.Stream == "" {
		ch.GM.Log.Error("stream event sent with no stream set")
		return
	}

	st, err := ch.GM.StreamManager.Find(schema.Stream)

	if err != nil {
		ch.GM.Log.Error(err)
		return
	}

	// If the stream is a broadcasted one, set the event to broadcast
	if st.Broadcast {
		e.Broadcast = true
	}

	// If we have a specific client id, set this on the event
	if schema.ClientId != "" {
		client, err := ch.GM.Server.Find(schema.ClientId)

		if err != nil {
			ch.GM.Log.Error("Invalid client id given: " + schema.ClientId)
			return
		}

		e.ClientId = schema.ClientId

		// Send the event to traits as well
		SendToTraits(client, e)
	}

	SendToClients(ch.GM, ch.Clients, e)

	// We also need to check if there are any handlers
	handlers, _ := ch.GM.StreamManager.FindHandlers(st.Name)

	for _, h := range handlers {
		h(schema.Data, st)
	}
}

func (ch *DirectChannel) Send(e Event, d EventDefinition) {
	if e.ClientId == "" {
		ch.GM.Log.Errorf("Direct event sent with no client id: %+v", e)
		return
	}

	client, err := ch.GM.Server.Find(e.ClientId)

	if err != nil {
		ch.GM.Log.Errorf("Direct event sent with unknown client: %s", e.ClientId)
		return
	}

	// Before sending directly to the client we should send this event
	// to any subscribers the client may have through its instanced components
	SendToTraits(client, e)
	client.Send <- e
}

// Send method for the internal channel
func (ch *InternalChannel) Send(e Event, d EventDefinition) {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			ch.GM.Log.Error(r)
		}
	}()

	subs, ok := ch.GM.Subscribers.Load(e.Name)

	if !ok {
		ch.GM.Log.Warningf("Internal event called with no active subscribers: %s", e.Name)
		return
	}

	subscribers := subs.([]EventHandler)

	// Fire all the things
	for _, sub := range subscribers {
		sub(e)
	}
}

// Connect on internal is useless
func (ch *InternalChannel) Connect(c *Client) {}

// Disconnect is also useless
func (ch *InternalChannel) Disconnect(c *Client) {}

// Send on the server channel can use the servers broadcast method
func (ch *ServerChannel) Send(e Event, d EventDefinition) {
	ch.GM.Server.Broadcast(e)
}

// Connect on server is useless
func (ch *ServerChannel) Connect(c *Client) {}

// Disconnect is also useless
func (ch *ServerChannel) Disconnect(c *Client) {}
