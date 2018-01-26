package engine

import (
    "errors"
    "reflect"
    "sync"
)

type (
    // Registry contains the reflect values for registered
    // structs, this allows us to register components and traits
    // once and dynamically load and interact with them without
    // the end user writing more code.
    Registry struct {
        Entries *sync.Map
    }

    // StructEntry contains details about a registered component/trait
    StructEntry struct {
        Name  string
        Value reflect.Type
    }
)

func NewRegistry() *Registry {
    return &Registry{
        Entries: new(sync.Map),
    }
}

// Load performs a bulk load into the registry
func (r *Registry) Load(m map[string]interface{}) {
    for k, v := range m {
        r.Entries.Store(k, &StructEntry{Name: k, Value: reflect.TypeOf(v)})
    }
}

// GetEntry returns the entry associated with the struct
func (r *Registry) GetEntry(n string) (*StructEntry, error) {
    e, k := r.Entries.Load(n)

    if !k {
        return nil, errors.New("Unable to find entry with the name " + n)
    }

    s := e.(*StructEntry)
    return s, nil
}

// GetStruct returns the underlying struct value of an entry
func (r *Registry) GetStruct(n string) (interface{}, error) {
    s, err := r.GetEntry(n)

    if err != nil {
        return nil, err
    }

    return reflect.New(s.Value).Elem().Interface(), nil
}
