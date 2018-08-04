package engine

import (
	"reflect"

	"github.com/caarlos0/env"
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

	// MongoSettings are used to denote how the mongo
	// wrapper interacts with the engine
	MongoSettings struct {
		Host        string `yaml:"host" env:"MONGO_HOST"`
		Database    string `yaml:"database" env:"MONGO_DATABASE"`
		AutoConnect bool   `yaml:"autoConnect" env:"MONGO_AUTOCONNECT"`
	}
)

// NewMongo creates a new instance of the wrapper
// if set, this will auto connect and fetch an
// instance of the mgo.Session.
func NewMongo(GM *GameManager) *Mongo {
	settings := MongoSettings{}
	env.Parse(&settings)

	mongo := &Mongo{GM: GM, Settings: settings}

	// If autoconnect has been set, create a new session
	if settings.AutoConnect {
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

// Save saves an entity and streams it back out
func (m *Mongo) Save(c string, i interface{}) {
	var bs bson.ObjectId
	entity := true
	id, ok := getField("ID", i)

	db, s := m.Instance()
	defer s.Close()

	if !ok {
		m.GM.Log.Warning("Couldn't find an id field, record is being inserted with no id.")
		entity = false
	}

	if ok && bson.IsObjectIdHex(id.(string)) {
		bs = bson.ObjectIdHex(id.(string))
	}

	if entity && bs.Valid() {
		q := bson.M{"entity._id": bs.Hex()}

		// Fire the On Update event
		ent := i.(EntityInterface)
		ent.OnUpdate(m.GM)

		// Update the record based on its id
		if err := db.C(c).Update(q, i); err != nil {
			m.GM.Log.Error(err)
			return
		}

	} else {

		if entity {
			// Update the entity fields
			ent := i.(EntityInterface)
			ent.OnCreate(bson.NewObjectId().Hex(), m.GM)
		}

		if err := db.C(c).Insert(i); err != nil {
			m.GM.Log.Error(err)
			return
		}
	}

	// If this is an entity, we want to send this to the stream manager
	if entity {
		m.GM.StreamManager.Updates(i)
	}
}

// Instance creates a copy of the session and returns that
func (m *Mongo) Instance() (*mgo.Database, *mgo.Session) {
	s := m.Session.Copy()
	return s.DB(m.Settings.Database), s
}

// Gets the value of a field if it exists
func getField(n string, i interface{}) (interface{}, bool) {
	re := reflect.ValueOf(i).Elem()
	if re.Kind() == reflect.Struct {
		f := re.FieldByName(n)

		if f.IsValid() {
			return f.Interface(), true
		}
	}

	return nil, false
}
