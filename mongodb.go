package main

import (
    "flag"
    "fmt"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "log"
    "strings"
)

var f_dbaddr = flag.String("mongo-address", "localhost:27017",
        "Address to the mongo database.")
var f_dbname = flag.String("mongo-db", "quickbump",
        "Database name to be used in MongoDb.")

func init() {
    RegisterDbImpl("mongodb", func() Database {
        db, err := NewMongoDb(*f_dbaddr, *f_dbname)
        if err != nil {
            log.Fatal(err)
        }
        return db
    })
}

func (id Id) toBsonId() (bson.ObjectId, error) {
    if bson.IsObjectIdHex(string(id)) {
        return bson.ObjectIdHex(string(id)), nil
    }
    return bson.ObjectId(""), fmt.Errorf("Invalid Id `%s'", id)
}

// Well, we just lost all the encapsulation we worked so hard to achieve ...
func (poll *PollData) SetBSON(raw bson.Raw) error {
    holder := struct{
        Mode PollMode
        Info bson.Raw
    }{}

    if err := raw.Unmarshal(&holder); err != nil {
        return err
    }
    poll.Mode = holder.Mode

    var info interface{}
    switch poll.Mode {
    case ChoicePoll:
        info = &ChoiceInfo{}
    case TextPoll:
        info = &TextInfo{}
    default:
        return UnknownPollMode(poll.Mode)
    }
    poll.Info = info
    return holder.Info.Unmarshal(info)
}


type MongoDb struct {
    session     *mgo.Session
    database    *mgo.Database
}

func NewMongoDb(location string, dbname string) (*MongoDb, error) {
    session, err := mgo.Dial(location)
    if err != nil {
        return nil, err
    }
    database := session.DB(dbname)
    return &MongoDb{session, database}, nil
}

func (db *MongoDb) Reset() error {
    return db.database.DropDatabase()
}

func (db *MongoDb) Close() {
    db.Close()
}

func (db *MongoDb) Insert(thing Thing) (Id, error) {
    bsonid := bson.NewObjectId()
    doc := &bson.M{
        "_id": bsonid,
        "value": thing,
    }
    c := db.database.C(Typename(thing))
    if err := c.Insert(doc); err != nil {
        return NullId, err
    }
    return Id(bsonid.Hex()), nil
}

func (db *MongoDb) Update(id Id, thing Thing) error {
    bsonid, err := id.toBsonId()
    if err != nil {
        return err
    }
    c := db.database.C(Typename(thing))
    if err := c.Update(
        &bson.M{"_id": bsonid},
        &bson.M{"value": thing}); err != nil {
        return err
    }
    return nil
}

func (db *MongoDb) Load(id Id, thing Thing) error  {
    bsonid, err := id.toBsonId()
    if err != nil {
        return err
    }
    c := db.database.C(Typename(thing))
    q := c.FindId(bsonid).Select(bson.M{"value": 1})
    holder := &struct{ Value bson.Raw }{}
    if err := q.One(holder); err != nil { // fixme
        return RecordNotFound
    }
    if err := bson.Unmarshal(holder.Value.Data, thing); err != nil {
        return err
    }
    return nil
}

func (db *MongoDb) Delete(id Id, thing Thing) error {
    bsonid, err := id.toBsonId()
    if err != nil {
        return err
    }
    c := db.database.C(Typename(thing))
    return c.RemoveId(bsonid)
}

func (db *MongoDb) Query(spec QuerySpec, thing Thing) Result {
    coll := db.database.C(Typename(thing))
    find := make(map[string]interface{})
    for name, values := range spec {
        if len(values) == 1 {
            fname := "value."+strings.ToLower(name)
            coll.EnsureIndexKey(fname)
            find[fname] = values[0] // ToLower because mgo magic ...
        }
        // todo, do something here
        //fmt.Errorf("Disjunctive query unsupported")
    }
    it := coll.Find(find).Iter()
    holder := &struct{
        Id      bson.ObjectId   "_id"
        Value   bson.Raw
    }{}
    return func(thing Thing) (Id, error) {
        if it.Next(&holder) {
            if err := bson.Unmarshal(holder.Value.Data, thing); err != nil {
                return NullId, err
            }
            return Id(holder.Id.Hex()), nil
        }
        err := it.Close()
        return NullId, err
    }
}
