package main

import (
	"github.com/yinhylin/goptions"
	"fmt"
	"github.com/BurntSushi/toml"
	"time"
)

type MapFoo struct {
	Key   string `cfg:"key"`
	Value string `cfg:"value"`
}

type StructFoo struct {
	Key   string `cfg:"key"`
	Value string `cfg:"value"`
}

type Any struct {
	TestAny string `cfg:"any-test"`
}

type MapTest struct {
	Any
	test         int
	TestInt      int               `cfg:"int"`
	TestInt8     int8              `cfg:"int8"`
	TestInt16    int16             `cfg:"int16"`
	TestInt32    int32             `cfg:"int32"`
	TestInt64    int64             `cfg:"int64"`
	TestDuration time.Duration     `cfg:"duration"`
	TestString   string            `cfg:"string"`
	TestBool     bool              `cfg:"bool"`
	SliceTest    []string          `cfg:"slice"`
	MapFoo       map[string]MapFoo `cfg:"map-foo"`
	StructFoo    StructFoo         `cfg:"struct-foo"`
}

func main() {
	var cfg map[string]interface{}
	opts := MapTest{
		MapFoo: make(map[string]MapFoo),
	}
	toml.DecodeFile("../testdata/test.toml", &cfg)
	fmt.Println(cfg)

	goptions.New(&opts).Merge("cfg", cfg)
	fmt.Println("result: ", &opts)
}
