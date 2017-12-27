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
    t.Event("tick", "fired for every tick of the game", true, []string{engine.INTERNAL_CHAN, engine.SERVER_CHAN})

    // Start the tick loop
    go t.Tick()
}

func (t *Ticker) Tick() {
    for {
        <-time.After(1 * time.Second)

        t.Count++
        t.Fire("tick", t)
    }
}
