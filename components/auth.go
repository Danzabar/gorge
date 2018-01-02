package components

import (
    "github.com/Danzabar/gorge/engine"
)

type (

    // AuthComponent deals with authorisation requests
    AuthComponent struct {
        engine.Component `json:"-"`
    }

    // User is the user entity
    User struct {
        engine.Entity
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    // Error is the event structure for a failed operation
    Error struct {
        Errors    map[string]string `json:"errors,omitempty"`
        Error     string            `json:"error,omitempty"`
        Operation string            `json:"operation"`
    }

    // Success is the event structure for a sucessful operation
    Success struct {
        User      *User  `json:"user"`
        Operation string `json:"operation"`
    }
)

func (a *AuthComponent) Register() {
    // Register events
    e.Event("auth:register", "event to register a new user", []string{engine.INTERNAL_CHAN})
    e.Event("auth:attempt", "event to attempt an authorisation", []string{engine.INTERNAL_CHAN})
    e.Event("auth:error", "event thrown when there was an error", []string{engine.DIRECT_CHAN})
    e.Event("auth:success", "event thrown when an auth event was processed", []string{engine.DIRECT_CHAN})

    // Bind to events
    e.Handler("auth:register", a.OnRegister)
    e.Handler("auth:attempt", a.OnAttempt)
}

func (a *AuthComponent) Setup() {
    // We can use this to migrate entities needed
    // for this component
    a.Migrate(&User{})
}

func (a *AuthComponent) OnRegister(e engine.Event) {

}

func (a *AuthComponent) OnAttempt(e engine.Event) {

}
