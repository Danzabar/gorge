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
