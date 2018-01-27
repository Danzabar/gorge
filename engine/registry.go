package engine

import (
    "errors"
    "reflect"
)

type (
    // Registry contains the reflect values for registered
    // structs, this allows us to register components and traits
    // once and dynamically load and interact with them without
    // the end user writing more code.
    Registry struct {
        Entries map[string]reflect.Type
    }
)

func NewRegistry() *Registry {
    return &Registry{
        Entries: make(map[string]reflect.Type),
    }
}

// Load performs a bulk load into the registry
func (r *Registry) Load(m map[string]interface{}) {
    for k, v := range m {
        r.Entries[k] = reflect.TypeOf(v)
    }
}

// GetStruct returns the underlying struct value of an entry
func (r *Registry) GetStruct(n string) (interface{}, error) {
    e, k := r.Entries[n]

    if !k {
        return nil, errors.New("Unable to find entry with the name " + n)
    }

    return reflect.New(e).Elem().Interface(), nil
}
