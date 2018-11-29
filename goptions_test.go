package goptions

import (
	"testing"
	"reflect"
	)

func TestNew(t *testing.T) {
	s := &struct {
	}{}
	vs := reflect.ValueOf(s)

	gop := New(s)
	if _, ok := interface{}(gop).(*Goptions); !ok {
		t.Errorf("get %T, want %T", gop, &Goptions{})
	}

	if gop.opts != vs {
		t.Errorf("get %T, want %T", gop.opts, vs)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("panic expected when *Goptions is not given.")
		}
	}()
	New(*s)
}

func TestMerge(t *testing.T) {
	s := &struct {
		TestString string `cfg:"test-string"`
		S struct {
			TestChild string `cfg:"test-child"`
		} `cfg:"s"`
	}{}
	cfg := make(map[string]interface{})
	cfg["test-string"] = "helloTest"
	child := make(map[string]interface{})
	child["test-child"] = "testChild"
	cfg["s"] = child

	gop := New(s).Merge("cfg", cfg)
	if _, ok := interface{}(gop).(*Goptions); !ok {
		t.Errorf("get %T, want %T", gop, &Goptions{})
	}

	if s.TestString != cfg["test-string"] {
		t.Errorf("get %v, want %v", s.TestString, cfg["test-string"])
	}

	if s.S.TestChild != child["test-child"] {
		t.Errorf("get %v, want %v", s.TestString, child["test-child"])
	}
}
