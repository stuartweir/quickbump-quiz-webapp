package main

import (
    "testing"
)

func TestMemDb(t *testing.T) {
    testDb(NewMemDb(), t)
}
