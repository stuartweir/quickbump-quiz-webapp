// Implements Database, stores everything in memory
// Data is actually encoded because we need to copy it and writing a
// deepcopy is hard :|

package main

import (
    "fmt"
    "log"
    "reflect"
    "encoding/json"
)

func init() {
    RegisterDbImpl("memdb", func() Database {
        return NewMemDb()
    })
}

type MemDb struct {
    id    uint
    store map[string]map[Id][]byte
}

func NewMemDb() (db *MemDb) {
    db = &MemDb{}
    db.Reset()
    return
}

func (m *MemDb) createId() Id {
    m.id += 1
    return Id(fmt.Sprintf("%x", m.id))
}

func (m *MemDb) Load(id Id, dst Thing) error {
    log.Printf("MemDb Loading %s : %s", id, dst)
    if data, ok := m.store[Typename(dst)][id]; ok {
        return json.Unmarshal(data, dst)
    }
    return RecordNotFound
}

func (m *MemDb) Insert(object Thing) (Id, error) {
    if object == nil {
        panic("Attempt to insert null into the database, this shouldn't happen, go find a programmer")
    }
    id := m.createId()
    data, err := json.Marshal(object)
    if err != nil {
        return "", err
    }
    name := Typename(object)
    if m.store[name] == nil {
        m.store[name] = make(map[Id][]byte)
    }
    m.store[name][id] = data
    log.Printf("MemDb Inserted %s : %s", id, object)
    return id, nil
}

func (m *MemDb) Update(id Id, object Thing) error {
    if _, found := m.store[Typename(object)][id]; !found {
        return RecordNotFound
    }
    if data, err := json.Marshal(object); err != nil {
        return err
    } else {
        m.store[Typename(object)][id] = data
        log.Printf("MemDb Updated %s : %s", id, object)
    }
    return nil
}

func (m *MemDb) Delete(id Id, object Thing) error {
    // Should we error if id isn't in our db?
    delete(m.store[Typename(object)], id)
    log.Printf("MemDb Deleted %s", id)
    return nil
}

func (m *MemDb) Close()  {
}

func (m *MemDb) Reset() error {
    m.id = 0
    m.store = make(map[string]map[Id][]byte)
    return nil
}

func (m *MemDb) Query(multiquery QuerySpec, derp Thing) Result {
    c := make(chan Id)
    errch := make(chan error)

    // This works by marshalling and umarshalling items in the query and
    // comparing them to unmarshaled items in the database. Yeah, it's
    // pretty dumb ...
    go func() {
        defer close(c)

        // Rebuild query
        query := make(map[string]interface{})
        for attr, valuelist := range multiquery {
            if len(valuelist) == 0 {
                continue
            }
            if len(valuelist) > 1 {
                errch <- fmt.Errorf("Disjunctive query unsupported")
                return
            }
            target := valuelist[0]
            if bytes, err := json.Marshal(&target); err != nil {
                errch <- fmt.Errorf("While evaluating query: %s", err)
                return
            } else {
                if err := json.Unmarshal(bytes, &target); err != nil {
                    errch <- fmt.Errorf("While evaluating query: %s", err)
                    return
                } else {
                    query[attr] = target
                }
            }
        }

        // Go find records
        modelholder := make(map[string]interface{})
        for id, data := range m.store[Typename(derp)] {
            if err := json.Unmarshal(data, &modelholder); err != nil {
                errch <- fmt.Errorf("While unmarshaling `%s': %s", data, err)
                return
            }

            ismatch := true
            for name, queryval := range query {
                if modelval, ok := modelholder[name]; !ok {
                    // Attribute in query not on model
                    errch <- fmt.Errorf("Field `%s' not found on `%s'", name, Typename(derp))
                    return
                } else {
                    if !reflect.DeepEqual(queryval, modelval) {
                        ismatch = false
                        break
                    }
                }
            }
            if ismatch {
                c <- id
            }
        }
    }()

    return func(thing Thing) (id Id, err error) {
        select {
        case err = <-errch:
            return *new(Id), err
        case id = <-c:
            if id == "" {
                return // don't try to load anything
            }
            err = m.Load(id, thing)
            return
        }
        panic("How did you get to this code path?")
    }
}
