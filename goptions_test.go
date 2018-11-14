package goptions

import (
	"testing"
)

type options struct {
	Foo bool `json:"foo" toml:"oof"`
}

func TestUse(t *testing.T) {
	opts := &options{}
	gop := Use(opts, "json", "flag")

	if _, ok := interface{}(gop).(*Goptions); !ok {
		t.Errorf("invalid *Goptions value type %T", gop)
	}

	if gop.tags[0] != "json" {
		t.Errorf("unexpected tag - %v, expected - %v", gop.tags[0], "json")
	}

	if gop.tags[1] != "flag" {
		t.Errorf("unexpected tag - %v, expected - %v", gop.tags[0], "flag")
	}
}

func TestMerge(t *testing.T) {
	opts := &options{Foo:false}
	cfg := make(map[string]interface{})
	cfg["foo"] = "true"
	gop := Use(opts, "json", "toml").Merge("json", cfg)

	if _, ok := interface{}(gop).(*Goptions); !ok {
		t.Errorf("invalid *Goptions value type %T", gop)
	}

	if !opts.Foo {
		t.Errorf("unexpected foo value - %v, expected - %v", opts.Foo, true)
	}

	cfg["oof"] = false
	gop.Merge("toml", cfg)
	if opts.Foo {
		t.Errorf("unexpected foo value - %v, expected - %v", opts.Foo, false)
	}
}
