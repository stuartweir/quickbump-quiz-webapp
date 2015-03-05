package main

import (
    "reflect"
    "sort"
    "testing"
    "time"
)

type Zebra struct {
    Name    string
}

type ZebraStripe struct {
    Zebra   Id
    Width   int
}

type ZebraAuction struct {
    Bids    []int
    Time    time.Time
    ZebraPtr  *Zebra
}

func testDb(db Database, t *testing.T) {
    testDbBasics(db, t)
    testDbLookup(db, t)
}

func testDbBasics(db Database, t *testing.T) {
    data_a := Zebra{"spam"}
    data_b := Zebra{"eggs"}

    {
    // Can't load what doesn't exist
    var v Zebra
    if err := db.Load("an id yet to be inserted", &v); err == nil {
        t.Fatal("Load didn't return an error when given an unused id")
    }

    // Can't update what doesn't exist
    var vv ZebraAuction
    if err := db.Update("an id yet to be inserted", &vv); err == nil {
        t.Fatal("Update didn't return an error when given an unused id")
    }
    }

    // Can insert
    id, err := db.Insert(&data_a)
    if err != nil {
        t.Fatal("Insert failed: ", err)
    }

    // Can load
    var v Zebra
    if err := db.Load(id, &v); err != nil {
        t.Fatal("Load failed: ", err)
    }
    if v != data_a {
        t.Fatalf("Result of Load did not match inserted value. Expected %s, Obtained %s", data_a, v)
    }

    // Can update
    if err := db.Update(id, &data_b); err != nil {
        t.Fatal("Update failed: ", err)
    }

    // Can load after update
    if err := db.Load(id, &v); err != nil {
        t.Fatal("Load failed: ", err)
    }
    if v != data_b {
        t.Fatalf("Result of Load did not match updated value. Expected %s, Obtained %s", data_b, v)
    }

    // Can Delete
    if err := db.Delete(id, &v); err != nil {
        t.Fatal("Delete failed: ", err)
    }

    // Can't load after Delete
    if err := db.Load(id, &v); err != RecordNotFound {
        t.Fatalf("Expected RecordNotFound. Err was %s", err)
    }
}

func testDbLookup(db Database, t *testing.T) {
    testDbLookupSimple(db, t)
    testDbLookupCompound(db, t)
}

func testDbLookupSimple(db Database, t *testing.T) {
    db.Insert(&ZebraAuction{
        []int{2,3,4},
        time.Now(),
        nil})

    derp := &ZebraAuction{}

    //{
    //result := db.Query(QuerySpec{"not an attribute": []interface{}{nil}}, derp)
    //_, err := result(derp)
    //if err == nil {
    //    t.Fatal("Expected error")
    //}
    //}

    {
    result := db.Query(QuerySpec{"ZebraPtr": []interface{}{nil}}, derp)
    id, err := result(derp)
    if id == "" {
        t.Fatalf("Expected result from query, id: %s, err: %v", id, err)
    }
    }
}

func testDbLookupCompound(db Database, t *testing.T) {
    zid, _ := db.Insert(&Zebra{"I am zebra"})
    stripes := []int{0, 1, 1, 2, 3, 5, 8}
    added := make([]string, len(stripes))
    for idx, v := range stripes {
        id, _ := db.Insert(ZebraStripe{zid, v})
        added[idx] = string(id)
    }

    model := &ZebraStripe{}
    result := db.Query(QuerySpec{"Zebra": []interface{}{zid}}, model)
    found := make([]string, len(stripes))
    for idx := 0;; idx++ {
        id, err := result(model)
        if err != nil {
            t.Fatal(err)
        }
        if id == "" {
            break
        }
        found[idx] = string(id)
    }

    sort.Strings(added)
    sort.Strings(found)
    if !reflect.DeepEqual(added, found) {
        t.Fatalf("%s != %s", added, found)
    }
}
