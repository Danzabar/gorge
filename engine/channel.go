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

    // InternalChannel works with component-component events
    InternalChannel struct {
        Channel
    }

    // ServerChannel is a channel that broadcasts by default
    ServerChannel struct {
        Channel
    }
)

func (ch *Channel) SetGM(GM *GameManager) {
    ch.GM = GM
}

// Open - On the base channel object, this isn't really needed
func (ch *Channel) Open() {
    ch.Clients = new(sync.Map)
}

// Send sends an event to clients on the channel
func (ch *Channel) Send(e Event, d EventDefinition) {
    // If the message is a broadcast send it to everyone
    if e.Broadcast {
        ch.Clients.Range(func(k, v interface{}) bool {
            client := v.(*Client)

            client.Send <- e
            return true
        })
    }

    if !e.Broadcast && e.ClientId == "" {
        ch.GM.Log.Errorf("Direct event sent with no client id: %+v", e)
        return
    }

    client, err := ch.GM.Server.Find(e.ClientId)

    if err != nil {
        ch.GM.Log.Errorf("Direct event sent with unknown client: %s", e.ClientId)
        return
    }

    client.Send <- e
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

// Send method for the internal channel
func (ch *InternalChannel) Send(e Event, d EventDefinition) {
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
