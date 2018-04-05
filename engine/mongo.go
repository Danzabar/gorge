package engine

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	// Mongo is the wrapper around mgo, it provides
	// some useful tools, hooks, validation flows etc.
	Mongo struct {
		GM       *GameManager
		Settings MongoSettings
		Session  *mgo.Session
	}

	Entity struct {
		Id        bson.ObjectId `bson:"_id,omitempty" json:"id,omitempty"`
		ClientId  string        `bson:"clientId" json:"clientId"`
		CreatedAt time.Time     `bson:"createdAt" json:"createdAt"`
		UpdatedAt time.Time     `bson:"updatedAt" json:"updatedAt"`
	}

	EntityInterface interface {
		// Used to control the collection name,
		GetCollection() string

		// Used to fetch the identifier for this record
		GetId() bson.ObjectId
		// Used to set the identifier for this record
		SetId(bson.ObjectId)
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
func NewMongo(GM *GameManager) *Mongo {
	// Since we have the GameManager we can
	// extract the values for mongo from
	// the global settings
	m := GM.Settings.Database.Mongo

	mongo := &Mongo{GM: GM, Settings: m}

	// If autoconnect has been set, create a new session
	if m.AutoConnect {
		mongo.Connect()
	}

	return mongo
}

// Connect attempts to connect to the server specified in config
func (m *Mongo) Connect() {
	sess, err := mgo.Dial(m.Settings.Host)

	if err != nil {
		panic(err)
	}

	sess.SetMode(mgo.Monotonic, true)
	m.Session = sess
	m.GM.Log.Info("Connected to mongo server...")
}

func (m *Mongo) Save(i EntityInterface) {
	n := i.GetCollection()
	id := i.GetId()

	if id.Valid() {
		// Update the record based on its id
		if err := m.Instance().C(n).UpdateId(id, i); err != nil {
			m.GM.Log.Error(err)
			return
		}
	} else {
		// Otherwise we can generate a new id and insert it
		i.SetId(bson.NewObjectId())

		if err := m.Instance().C(n).Insert(&i); err != nil {
			m.GM.Log.Error(err)
			return
		}
	}

	// TODO inform the stream manager that entity has changed
}

// Instance creates a copy of the session and returns that
func (m *Mongo) Instance() *mgo.Database {
	s := m.Session.Copy()
	return s.DB(m.Settings.Database)
}

// GetId default for mongo entities
func (e *Entity) GetId() bson.ObjectId {
	return e.Id
}

func (e *Entity) SetId(b bson.ObjectId) {
	e.Id = b
}
