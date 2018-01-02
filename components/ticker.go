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

// Register registers the events used by this component
func (t *Ticker) Register() {
    t.Event("tick", "fired every second as a marker for other components", []string{engine.INTERNAL_CHAN, engine.SERVER_CHAN})
}

// Run is a method that is called when the game has started
func (t *Ticker) Run() {
    t.Tick()
}

func (t *Ticker) Tick() {
    for {
        <-time.After(1 * time.Second)

        t.Count++
        t.Fire("tick", t)
    }
}
