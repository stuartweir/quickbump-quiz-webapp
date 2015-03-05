package main

import (
    "errors"
    "reflect"
    "fmt"
)

var RecordNotFound = errors.New("Record not found")

type Id string
var NullId Id = *new(Id)

type Thing interface{
}

type Validatable interface {
    Validate() error
}
type Database interface {
    Insert(Thing) (Id, error)
    Update(Id, Thing) error
    Load(Id, Thing) error
    // todo, deprecate 2nd arg in Delete and Query
    Delete(Id, Thing) error
    Query(QuerySpec, Thing) Result
    Reset() error
    Close()
}

// Calling a result loads the next result into Thing and returns its id
type Result func(Thing) (Id, error)
type QuerySpec map[string][]interface{}

// ...
func Typename(thingy interface{}) (name string) {
    tipe := reflect.TypeOf(thingy)
    if tipe == nil {
        panic("Oh boy ...")
    }
    if tipe.Kind() == reflect.Ptr {
        tipe = tipe.Elem()
    }
    name = tipe.Name()
    if name == "" {
        panic(fmt.Sprintf("empty typename for `%s'", thingy))
    }
    return
}

type DbFactory func() Database
var DbRegistry = make(map[string]DbFactory)

// Register a factory which returns a Database implementation
func RegisterDbImpl(name string, factory DbFactory) {
    if _, exists := DbRegistry[name]; exists {
        e := fmt.Sprintf("Unable to register database factory %v named %s." +
                "Factory already registered under that name.", factory, name)
        panic(e)
    }
    DbRegistry[name] = factory
}
