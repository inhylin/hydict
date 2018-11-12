package goptions

import (
	"testing"
)

type options struct {
	Foo string `json:"foo"`
}

func TestUse(t *testing.T) {
	opts := &options{}
	gop := Use(opts, []string{"json"})
	if gop.tags[0] != "json" {
		t.Fatalf("unexpected tag - %s, expected - %s", gop.tags[0], "json")
	}
}
