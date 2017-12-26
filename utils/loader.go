package utils

import (
    "github.com/Danzabar/gorge/components"
    "github.com/Danzabar/gorge/engine"
)

// LoadDefaultComponents loads the standard components
func LoadDefaultComponents(GM *engine.GameManager) {
    GM.AddComponents(map[string]engine.ComponentInterface{
        "ticker": &components.Ticker{},
    })
}
