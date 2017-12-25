package engine

import (
    "sync"

    "github.com/gorilla/websocket"
)

type (

    // Client represents a connected client/user
    Client struct {
        Id   string `json:"id"`
        Conn *websocket.Conn
        Send chan Event
    }

    // Server represents the collection of connected
    // clients
    Server struct {
        GM         *GameManager
        Clients    *sync.Map
        Channels   *sync.Map
        Register   chan *Client
        Disconnect chan *Client
    }
)

// NewServer creates a new instance of the server struct
func NewServer() *Server {
    return &Server{
        Clients:    new(sync.Map),
        Channels:   new(sync.Map),
        Register:   make(chan *Client),
        Disconnect: make(chan *Client),
    }
}

// Connect adds a new client to the server
func (s *Server) Connect(client *Client) {
    s.Clients.Store(client.Id, client)

    go client.Reader()
    go client.Writer()
}

// Disconnect removes a client from the server
func (s *Server) Disconnect(client *Client) {
    s.Clients.Delete(client.Id)
    close(client.Send)
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
                s.Disconnect <- c
                break readloop
            }
        }

        // Set the info we already know about the event
        e.Inbound = true
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
        case u := <-s.Disconnect:
            s.Disconnect(u)
        }
    }
}
