package engine

import (
    "errors"
    "sync"

    "github.com/gorilla/websocket"
)

type (

    // Client represents a connected client/user
    Client struct {
        Id   string          `json:"id"`
        Conn *websocket.Conn `json:"-"`
        Send chan Event      `json:"-"`
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
    GM.Event(EventDefinition{"connected", "fired when a new client connects", []string{INTERNAL_CHAN}})
    GM.Event(EventDefinition{"disconnect", "fired when a client disconnects", []string{INTERNAL_CHAN}})

    // Add the default channels
    serv.NewChannels(map[string]ChannelInterface{
        INTERNAL_CHAN: &InternalChannel{},
        DIRECT_CHAN:   &Channel{},
        SERVER_CHAN:   &ServerChannel{},
    })

    go serv.Listen()
    return serv
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

    go client.Reader(s)
    go client.Writer(s)

    s.GM.FireEvent(NewEvent("connected", client))
    s.GM.Log.Debug("New client connected")
}

// Disconnect removes a client from the server
func (s *Server) Disconnect(client *Client) {
    s.Clients.Delete(client.Id)
    close(client.Send)

    s.GM.FireEvent(NewEvent("disconnected", client))
    s.GM.Log.Debug("Client disconnected")
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
func (c *Client) Reader(s *Server) {
readloop:
    for {
        var e Event

        if err := c.Conn.ReadJSON(&e); err != nil {

            // Catch for client disconnects
            if _, k := err.(*websocket.CloseError); k {
                s.Unregister <- c
                break readloop
            }
        }

        // Set the info we already know about the event
        e.ClientId = c.Id

        // Finally fire the event
        s.GM.FireEvent(e)
    }
}

// Writer writes messages to the given client
func (c *Client) Writer(s *Server) {
writerloop:
    for {
        select {
        case event, ok := <-c.Send:

            if !ok {
                // At this point, we would assume a disconnection
                break writerloop
            }

            if err := c.Conn.WriteJSON(event); err != nil {
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
