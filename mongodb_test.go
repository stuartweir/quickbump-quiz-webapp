package main

import (
    "testing"
)

func TestMongoDb(t *testing.T) {
    if db, err := NewMongoDb("localhost:27017", "qbtest"); err != nil {
        t.Fatal(err)
    } else {
        // Don't try to call this, it will mess everything up ...
        //defer db.Close()
        db.Reset()
        testDb(db, t)
    }
}
