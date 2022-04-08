package main

import (
	"reflect"
	"testing"
)

func TestGetEventHeader(t *testing.T) {
	expect := []string{"seq", "ts", "what"}
	if got := GetEventColumnNames(); !reflect.DeepEqual(expect, got) {
		t.Fatalf("Expected: %q, got: %q", expect, got)
	}
}
