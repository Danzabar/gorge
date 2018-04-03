package engine

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

type (

	// Client represents a connected client/user
	Client struct {
		Id          string              `json:"id"`
		Conn        ConnectionInterface `json:"-"`
		Send        chan Event          `json:"-"`
		Traits      *sync.Map           `json:"-"`
		Subscribers *sync.Map           `json:"-"`
	}

	// ConnectionInterface defines what we expect from a connection
	ConnectionInterface interface {
		Reader(c *Client, s *Server)
		Writer(c *Client, s *Server)
	}

	// Server represents the collection of connected
	// clients
	Server struct {
		GM         *GameManager
		Clients    *sync.Map
		Channels   *sync.Map
		Register   chan *Client
		Unregister chan *Client
		Shutdown   chan bool
	}

	// WebsocketConnection is the default connection used
	// provides a websocket reader and writer for the client to
	// connect to
	WebsocketConnection struct {
		Conn *websocket.Conn
	}
)

// NewServer creates a new instance of the server struct
func NewServer(GM *GameManager) *Server {
	serv := &Server{
		GM:         GM,
		Clients:    new(sync.Map),
		Channels:   new(sync.Map),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	// Register events
	GM.Event(EventDefinition{"connected", "", false, false, []string{INTERNAL_CHAN, DIRECT_CHAN}})
	GM.Event(EventDefinition{"disconnect", "", false, false, []string{INTERNAL_CHAN, DIRECT_CHAN}})

	// Add the default channels
	serv.NewChannels(map[string]ChannelInterface{
		INTERNAL_CHAN: &InternalChannel{},
		DIRECT_CHAN:   &DirectChannel{},
		SERVER_CHAN:   &ServerChannel{},
	})

	return serv
}

// NewClient creates a new client from the given details
func NewClient(c ConnectionInterface, id string) *Client {
	return &Client{
		Id:          id,
		Conn:        c,
		Send:        make(chan Event),
		Traits:      new(sync.Map),
		Subscribers: new(sync.Map),
	}
}

// RegisterHandler registers a handler for an Instanced component
// all handlers for instances will bind to the `direct` channel
// this means they will be presented with all personalised events
func (c *Client) RegisterHandler(n string, h EventHandler) {
	var handlers []EventHandler

	reg, ok := c.Subscribers.Load(n)

	if ok {
		handlers = reg.([]EventHandler)
	}

	handlers = append(handlers, h)
	c.Subscribers.Store(n, handlers)
}

// RemoveTrait removes the trait for the clients list of traits
// and triggers the destroy event method on the trait interface
func (c *Client) RemoveTrait(n string, inst *Instance) {
	t, k := c.Traits.Load(n)

	if !k {
		// TODO: add a log here once clients have access to it
		return
	}

	// other wise we can destroy the trait
	trait := t.(TraitInterface)

	trait.Destroy(inst)
	c.Traits.Delete(n)
}

// BindInstance adds a new instance to the client
func (c *Client) BindTrait(n string, i TraitInterface, inst *Instance) {
	// We don't really care if the instance already exists
	// we can replace it with the provided instance
	inst.SetClient(c)

	// Add to store
	c.Traits.Store(n, i)

	// Fire the connect event
	i.Connect(inst)
}

// Forward allows a channel to forward an event even if the event
// was never meant for the given channel
func (s *Server) Forward(n string, e Event, d EventDefinition) {
	// Does the channel exist?
	ch, err := s.FindChannel(n)

	if err != nil {
		s.GM.Log.Errorf("attempt to forward an event to a channel that doesn't exist: %s", n)
		return
	}

	// if so, forward the event
	ch.Send(e, d)
}

// SendToChannels uses the channels on an event definition to send
// the events to right clients
func (s *Server) SendToChannels(e Event, d EventDefinition) {
	for _, v := range d.Channels {
		ch, err := s.FindChannel(v)

		if err != nil {
			s.GM.Log.Errorf("Unable to find channel %s - Cannot send event %+v", v, e)
			continue
		}

		ch.Send(e, d)
	}
}

// NewChannels creates and adds channels to the store
func (s *Server) NewChannels(c map[string]ChannelInterface) {
	for k, v := range c {
		v.SetGM(s.GM)
		v.Open()
		s.Channels.Store(k, v)
	}
}

// ConnectTo connects the given client to a channel by name
func (s *Server) ConnectTo(n string, c *Client) {
	ch, err := s.FindChannel(n)

	if err != nil {
		s.GM.Log.Errorf("Unable to connect to channel %s - it has not yet been registered", n)
		return
	}

	ch.Connect(c)
}

// FindChannel attempts to fetch a channel from the store
func (s *Server) FindChannel(n string) (ChannelInterface, error) {
	ch, ok := s.Channels.Load(n)

	if !ok {
		return &Channel{}, errors.New("channel " + n + " does not exist")
	}

	return ch.(ChannelInterface), nil
}

// Find attempts to get a client by its identifier
func (s *Server) Find(id string) (*Client, error) {
	cl, ok := s.Clients.Load(id)

	if !ok {
		return &Client{}, errors.New("client doesn't exist")
	}

	return cl.(*Client), nil
}

// Connect adds a new client to the server
func (s *Server) Connect(client *Client) {
	s.Clients.Store(client.Id, client)

	s.GM.Log.Infof("Connecting new client %s", client.Id)

	go client.Conn.Reader(client, s)
	go client.Conn.Writer(client, s)

	s.GM.FireEvent(NewDirectEvent("connected", client, client.Id))
}

// Disconnect removes a client from the server
func (s *Server) Disconnect(client *Client) {
	s.Clients.Delete(client.Id)
	close(client.Send)

	s.GM.FireEvent(NewDirectEvent("disconnected", client, client.Id))
}

// Broadcast sends a message to all connected clients
func (s *Server) Broadcast(e Event) {
	s.Clients.Range(func(k, v interface{}) bool {
		client := v.(*Client)
		client.Send <- e
		return true
	})
}

// Reader reads messages from the client and processess
// them as events
func (ws *WebsocketConnection) Reader(c *Client, s *Server) {
readloop:
	for {
		var e Event

		if err := ws.Conn.ReadJSON(&e); err != nil {

			s.GM.Log.Error(err)

			// Catch for client disconnects
			if _, k := err.(*websocket.CloseError); k {
				s.Unregister <- c
				break readloop
			}
		}

		// Set the info we already know about the event
		e.ClientId = c.Id
		// We also know this was of the inbound origin
		e.Origin = ORIG_CLIENT

		// Finally fire the event
		s.GM.FireEvent(e)
	}
}

// Writer writes messages to the given client
func (ws *WebsocketConnection) Writer(c *Client, s *Server) {
writerloop:
	for {
		select {
		case event, ok := <-c.Send:

			if !ok {
				// At this point, we would assume a disconnection
				break writerloop
			}

			if err := ws.Conn.WriteJSON(event); err != nil {
				s.GM.Log.Errorf("Unable to process event: %+v", event)
			}
		}
	}
}

// Listen starts the server loop
func (s *Server) Listen() {
servloop:
	for {
		select {
		case r := <-s.Register:
			s.Connect(r)
		case u := <-s.Unregister:
			s.Disconnect(u)
		case <-s.Shutdown:
			break servloop
		}
	}
}
