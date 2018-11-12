package goptions

import (
	"reflect"
)

type Goptions struct {
	tags        []string
	value       *reflect.Value
	TagMappings map[string]TagMapping
}

type TagMapping map[int]Mapping

type Mapping struct {
	Name    string
	TagName string
	Value   *reflect.Value
}

func Use(opts interface{}, tags []string) *Goptions {
	gop := &Goptions{
		tags:        tags,
		TagMappings: make(map[string]TagMapping),
	}
	resolve(gop, opts)
	return gop
}

func resolve(gop *Goptions, opts interface{}) {
	val := reflect.ValueOf(opts).Elem()
	gop.value = &val

	for _, tag := range gop.tags {
		gop.TagMappings[tag] = make(TagMapping)
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.FieldByName(field.Name)

		for _, tag := range gop.tags {
			tagName := field.Tag.Get(tag)
			if tagName == "" {
				delete(gop.TagMappings[tag], i)
				continue
			}

			gop.TagMappings[tag][i] = Mapping{
				Name:    field.Name,
				TagName: tagName,
				Value:   &fieldVal,
			}
		}
	}
}
