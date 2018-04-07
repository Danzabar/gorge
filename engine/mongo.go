package engine

import (
	"reflect"
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
		SetID(bson.ObjectId)
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

func (m *Mongo) Save(c string, i interface{}) {
	var nonEntity bool
	var bs bson.ObjectId

	id, ok := m.getField("Id", i)

	if !ok {
		m.GM.Log.Warning("Couldn't find an id field, record is being inserted with no id.")
		nonEntity = true
	} else {
		bs = id.(bson.ObjectId)
	}

	if !nonEntity && bs.Valid() {

		// Update the record based on its id
		if err := m.Instance().C(c).UpdateId(bs, i); err != nil {
			m.GM.Log.Error(err)
			return
		}

	} else {

		if !nonEntity {
			// Update the entity fields
			ent := i.(EntityInterface)
			ent.SetID(bson.NewObjectId())
		}

		if err := m.Instance().C(c).Insert(i); err != nil {
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

func (m *Mongo) getField(n string, i interface{}) (interface{}, bool) {
	re := reflect.ValueOf(i).Elem()

	if re.Kind() == reflect.Struct {
		f := re.FieldByName(n)

		if f.IsValid() {
			return f.Interface(), true
		}
	}

	return nil, false
}

func (e *Entity) SetID(in bson.ObjectId) {
	e.Id = in
}
