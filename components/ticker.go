package components

import (
    "time"

    "github.com/Danzabar/gorge/engine"
)

type (
    Ticker struct {
        engine.Component `json:"-"`
        Count            int `json:"count"`
    }
)

func (t *Ticker) Register() {
    t.Event("tick", "fired for every tick of the game", false, false, true)

    // Start the tick loop
    go t.Tick()
}

func (t *Ticker) Tick() {
    <-time.After(1 * time.Second)

    c.Fire("tick", t)
}
