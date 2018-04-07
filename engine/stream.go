package engine

import (
	"errors"
	"reflect"
	"sync"
)

const (
	// Constant for the stream channel name
	STREAM_CHAN = "stream"

	STREAM_SAVE_EVENT    = "stream.save"
	STREAM_UPDATED_EVENT = "stream.updated"
)

type (
	StreamManager struct {
		GM             *GameManager
		Streams        *sync.Map
		StreamHandlers *sync.Map
	}

	Stream struct {
		Name        string
		Collection  string
		StructValue reflect.Type
		Broadcast   bool
	}

	StreamSchema struct {
		Stream   string      `json:"stream"`
		ClientId string      `json:"-"`
		Data     interface{} `json:"data"`
	}

	// StreamHandler defines the expectations of a handler
	// for streaming events
	StreamHandler func(i interface{}, s *Stream)
)

func NewStreamManager(GM *GameManager) *StreamManager {
	return &StreamManager{
		GM:             GM,
		Streams:        new(sync.Map),
		StreamHandlers: new(sync.Map),
	}
}

func (s *StreamManager) Register() {
	// Register the events
	s.registerEvents()

	// Register channel
	s.GM.Server.NewChannels(map[string]ChannelInterface{STREAM_CHAN: &StreamChannel{}})

	// Register event handlers
	s.GM.RegisterHandler(STREAM_SAVE_EVENT, s.OnSave)
	s.GM.RegisterHandler("connected", s.OnConnect)
}

// Registers the events used by the stream component
func (s *StreamManager) registerEvents() {
	s.GM.Event(EventDefinition{Name: STREAM_SAVE_EVENT, StrictSchema: false, TrustExternal: true, Channels: []string{INTERNAL_CHAN}})
	s.GM.Event(EventDefinition{Name: STREAM_UPDATED_EVENT, StrictSchema: false, TrustExternal: false, Channels: []string{STREAM_CHAN}})
}

// New creates a new Stream object and adds it to the store
func (s *StreamManager) New(n string, c string, i interface{}, b bool) {
	st := &Stream{
		Name:        n,
		Collection:  c,
		StructValue: reflect.TypeOf(i),
		Broadcast:   b,
	}

	s.Streams.Store(st.Name, st)
}

// Updates is used to tell stream manager when an entity
// has been updated/saved
func (s *StreamManager) Updates(i interface{}) {
	var stream *Stream

	schema := &StreamSchema{Data: i}
	rt := reflect.TypeOf(i).Elem()

	s.Streams.Range(func(k, v interface{}) bool {
		st := v.(*Stream)

		if st.StructValue.Name() == rt.Name() {
			stream = st
			return false
		}

		return true
	})

	if stream == nil {
		return
	}

	cl, ok := getField("ClientId", i)

	if ok {
		schema.ClientId = cl.(string)
	}

	schema.Stream = stream.Name
	s.GM.FireEvent(NewDirectEvent(STREAM_UPDATED_EVENT, schema, schema.ClientId))
}

// FindHandlers finds handlers with the given stream name
func (s *StreamManager) FindHandlers(n string) ([]StreamHandler, error) {
	sh, ok := s.StreamHandlers.Load(n)

	if !ok {
		return []StreamHandler{}, errors.New("Unable to locate any stream handlers for " + n)
	}

	return sh.([]StreamHandler), nil
}

// Handler registers a new stream handler
func (s *StreamManager) Handler(n string, h StreamHandler) {
	st, err := s.FindHandlers(n)

	if err != nil {
		st = []StreamHandler{}
	}

	st = append(st, h)
	s.StreamHandlers.Store(n, st)
}

// Handler for save events
func (s *StreamManager) OnSave(e Event) bool {
	var schema StreamSchema

	Decode(e.Data, &schema)

	// Find the stream
	st, err := s.Find(schema.Stream)

	if err != nil {
		s.GM.Log.Error(err)
		return false
	}

	val := reflect.New(st.StructValue).Interface()
	Decode(schema.Data, val)

	// If this is an entity, we should set the client id
	entity, ok := val.(EntityInterface)

	if ok {
		entity.SetClientId(e.ClientId)
	}

	// Save the data
	s.GM.DB.Save(st.Collection, val)

	return true
}

// Find does what it says on the tin
func (s *StreamManager) Find(n string) (*Stream, error) {
	st, ok := s.Streams.Load(n)

	if !ok {
		return &Stream{}, errors.New("stream " + n + " does not exist")
	}

	return st.(*Stream), nil
}

// OnConnect is an event handler fired when
// a client connects to the server. In this
// instance we are connecting them to the streaming
// channel
func (s *StreamManager) OnConnect(e Event) bool {
	cl, err := s.GM.Server.Find(e.ClientId)

	if err != nil {
		s.GM.Log.Error(err)
		return false
	}

	s.GM.Server.ConnectTo(STREAM_CHAN, cl)
	return true
}
