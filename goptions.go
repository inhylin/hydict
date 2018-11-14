package goptions

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Goptions struct {
	tags        []string
	value       *reflect.Value
	TagMappings map[string]TagMapping // map[tag]TagMapping
}

type TagMapping []Mapping

type Mapping struct {
	Name    string
	TagName string
	Value   *reflect.Value
}

func Use(opts interface{}, tags ...string) *Goptions {
	gop := &Goptions{
		tags:        tags,
		TagMappings: make(map[string]TagMapping),
	}
	gop.resolve(opts)
	return gop
}

func (gop *Goptions) Merge(tag string, cfg map[string]interface{}) *Goptions {
	if ms, ok := gop.TagMappings[tag]; ok {
		for _, m := range ms {
			if val, ok := cfg[m.TagName]; ok {
				gop.merge(m, val)
			}
		}
	} else {
		log.Fatalf("unknown tag %s, please Use it before Merge", tag)
	}

	return gop
}

func (gop *Goptions) MergeFlag(flagSet *flag.FlagSet) *Goptions {
	tag := "flag"
	if ms, ok := gop.TagMappings[tag]; ok {
		for _, m := range ms {
			f := flagSet.Lookup(m.TagName)
			if f != nil {
				gop.merge(m, f.Value.String())
			}
		}
	}

	return gop
}

func (gop *Goptions) resolve(opts interface{}) {
	val := reflect.ValueOf(opts).Elem()
	gop.value = &val

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.FieldByName(field.Name)

		if field.Anonymous {
			var fieldPtr reflect.Value
			switch val.FieldByName(field.Name).Kind() {
			case reflect.Struct:
				fieldPtr = val.FieldByName(field.Name).Addr()
			case reflect.Ptr:
				fieldPtr = reflect.Indirect(val).FieldByName(field.Name)
			}
			if !fieldPtr.IsNil() {
				gop.resolve(fieldPtr.Interface())
			}
		}

		for _, tag := range gop.tags {
			tagName := field.Tag.Get(tag)
			if tagName == "" {
				continue
			}

			gop.TagMappings[tag] = append(gop.TagMappings[tag], Mapping{
				Name:    field.Name,
				TagName: tagName,
				Value:   &fieldVal,
			})
		}
	}
}

func (gop *Goptions) merge(m Mapping, v interface{}) {
	formatted, err := format(m, v)
	if err != nil {
		log.Fatalf("option resolution failed to format %v for %s (%+v) - %s",
			v, m.Name, m.Value, err)
	}

	m.Value.Set(reflect.ValueOf(formatted))
}

func format(m Mapping, v interface{}) (interface{}, error) {
	switch m.Value.Interface().(type) {
	case bool:
		return formatBool(v)
	case int:
		i, err := formatInt64(v)
		if err != nil {
			return nil, err
		}
		return int(i), nil
	case int16:
		i, err := formatInt64(v)
		if err != nil {
			return nil, err
		}
		return int16(i), nil
	case int32:
		i, err := formatInt64(v)
		if err != nil {
			return nil, err
		}
		return int32(i), nil
	case int64:
		return formatInt64(v)
	case uint:
		i, err := formatInt64(v)
		if err != nil {
			return nil, err
		}
		return uint(i), nil
	case uint16:
		i, err := formatInt64(v)
		if err != nil {
			return nil, err
		}
		return uint16(i), nil
	case uint32:
		i, err := formatInt64(v)
		if err != nil {
			return nil, err
		}
		return uint32(i), nil
	case uint64:
		i, err := formatInt64(v)
		if err != nil {
			return nil, err
		}
		return uint64(i), nil
	case float32:
		i, err := formatFloat64(v)
		if err != nil {
			return nil, err
		}
		return float32(i), nil
	case []float32:
		return formatFloat32Slice(v)
	case float64:
		return formatFloat64(v)
	case []float64:
		return formatFloat64Slice(v)
	case string:
		return formatString(v)
	case []string:
		return formatStringSlice(v)
	case time.Duration:
		return formatDuration(v)
	}
	return nil, nil
}

func formatBool(v interface{}) (bool, error) {
	switch v.(type) {
	case bool:
		return v.(bool), nil
	case string:
		return strconv.ParseBool(v.(string))
	case int, int16, uint16, int32, uint32, int64, uint64:
		return reflect.ValueOf(v).Int() == 0, nil
	}
	return false, fmt.Errorf("invalid bool value type %T", v)
}

func formatInt64(v interface{}) (int64, error) {
	switch v.(type) {
	case int, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint16, uint32, uint64:
		return int64(reflect.ValueOf(v).Uint()), nil
	case string:
		return strconv.ParseInt(v.(string), 10, 64)
	}

	return 0, fmt.Errorf("invalid int64 value type %T", v)
}

func formatFloat64(v interface{}) (float64, error) {
	switch v.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float(), nil
	case string:
		return strconv.ParseFloat(v.(string), 64)
	}

	i64, err := formatInt64(v)
	if err == nil {
		return float64(i64), nil
	}

	return 0, fmt.Errorf("invalid float64 value type %T", v)
}

func formatDuration(v interface{}) (time.Duration, error) {
	switch v.(type) {
	case time.Duration:
		return v.(time.Duration), nil
	case int, int16, int32, int64, uint, uint16, uint32, uint64:
		return time.Duration(reflect.ValueOf(v).Int()) * time.Millisecond, nil
	case string:
		if regexp.MustCompile(`^[0-9]+$`).MatchString(v.(string)) {
			intVal, err := strconv.Atoi(v.(string))
			if err != nil {
				return 0, err
			}
			return time.Duration(intVal) * time.Millisecond, nil
		}
		return time.ParseDuration(v.(string))
	}

	return 0, fmt.Errorf("invalid time.Duration value type %T", v)
}

func formatString(v interface{}) (string, error) {
	return fmt.Sprintf("%s", v), nil
}

func formatStringSlice(v interface{}) ([]string, error) {
	var tmp []string
	switch v.(type) {
	case []string:
		tmp = v.([]string)
	case []interface{}:
		for _, ss := range v.([]interface{}) {
			tmp = append(tmp, ss.(string))
		}
	case string:
		for _, s := range strings.Split(v.(string), ",") {
			tmp = append(tmp, s)
		}
	}
	return tmp, nil
}

func formatFloat32Slice(v interface{}) ([]float32, error) {
	var tmp []float32
	switch v.(type) {
	case []float32:
		tmp = v.([]float32)
	case []interface{}:
		for _, i := range v.([]interface{}) {
			tmp = append(tmp, i.(float32))
		}
	case string:
		for _, s := range strings.Split(v.(string), ",") {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, float32(f))
		}
	case []string:
		for _, s := range v.([]string) {
			f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, float32(f))
		}
	}
	return tmp, nil
}

func formatFloat64Slice(v interface{}) ([]float64, error) {
	var tmp []float64
	switch v.(type) {
	case []float64:
		tmp = v.([]float64)
	case []interface{}:
		for _, i := range v.([]interface{}) {
			tmp = append(tmp, i.(float64))
		}
	case string:
		for _, s := range strings.Split(v.(string), ",") {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, f)
		}
	case []string:
		for _, s := range v.([]string) {
			f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, f)
		}
	}
	return tmp, nil
}
