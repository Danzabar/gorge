package engine

import (
    "gopkg.in/mgo.v2"
)

type (
    // Mongo is the wrapper around mgo, it provides
    // some useful tools, hooks, validation flows etc.
    Mongo struct {
        GM       *GameManager
        Settings MongoSettings
        Session  *mgo.Session
    }

    // MongoSettings are used to denote how the mongo
    // wrapper interacts with the engine
    MongoSettings struct {
        Host        string `yaml:"host"`
        Database    string `yaml:"database"`
        AutoSave    bool   `yaml:"autoSave"`
        AutoConnect bool   `yaml:"autoConnect"`
    }
)

// NewMongo creates a new instance of the wrapper
// if set, this will auto connect and fetch an
// instance of the mgo.Session.
//
// A note on this: Gorge will not work currently
// without this wrapper, it is there to provide
// extended functionality to mgo, it does not replace
// any functionality.
func NewMongo(GM *GameManager, m MongoSettings) *Mongo {
    mongo := &Mongo{GM: GM, Settings: m}

    // If autoconnect has been set, create a new session
    if m.AutoConnect {
        sess, err := mgo.Dial(m.Host)

        // We panic here because there is no point
        // starting the service if we expect storage and
        // it isn't there.
        if err != nil {
            panic(err)
        }

        mongo.Session = sess
    }

    return mongo
}
