package goptions

import (
	"reflect"
	"fmt"
	"strconv"
	"time"
	"regexp"
	"flag"
)

type Goptions struct {
	currentTag string
	opts       reflect.Value
}

func New(opts interface{}) *Goptions {
	gop := &Goptions{}

	vo := reflect.ValueOf(opts)
	if vo.Kind() != reflect.Ptr || vo.Elem().Kind() != reflect.Struct {
		panic("The expected opts type is Pointer of Struct, " + vo.Kind().String() + " is given.")
	}
	gop.opts = vo

	return gop
}

func (gop *Goptions) Merge(tag string, cfg map[string]interface{}) *Goptions {
	gop.currentTag = tag

	gop.merge(gop.opts, cfg)

	return gop
}

func (gop *Goptions) MergeFlag(flagSet *flag.FlagSet) *Goptions {
	cfg := make(map[string]interface{})

	flagSet.Visit(func(i *flag.Flag) {
		cfg[i.Name] = i.Value.String()
	})

	gop.currentTag = "flag"
	gop.merge(gop.opts, cfg)

	return gop
}

func (gop *Goptions) merge(vv reflect.Value, cfg map[string]interface{}) {
	ve := vv
	if vv.Kind() == reflect.Ptr {
		ve = vv.Elem()
	}
	te := ve.Type()

	for i := 0; i < te.NumField(); i++ {
		key := te.Field(i)
		val := ve.FieldByName(key.Name)

		if val.CanSet() {
			if key.Anonymous {
				gop.merge(val, cfg)
			}

			tag := key.Tag.Get(gop.currentTag)
			if tag != "" {
				if c, ok := cfg[tag]; ok {
					gop.resolve(val, c)
				}
			}
		}
	}
}

func (gop *Goptions) resolve(vv reflect.Value, cfg interface{}) {
	ve := vv
	if vv.Kind() == reflect.Ptr {
		ve = vv.Elem()
	}
	te := ve.Type()

	switch ve.Kind() {
	case reflect.Ptr:
		gop.resolve(ve.Elem(), cfg)
	case reflect.Map:
		cfg := cfg.(map[string]interface{})
		vm := reflect.MakeMap(te)
		for kc, vc := range cfg {
			key := reflect.New(te.Key())
			gop.resolve(key, kc)
			val := reflect.New(te.Elem())
			gop.resolve(val, vc)

			vm.SetMapIndex(key.Elem(), val.Elem())
		}
		ve.Set(vm)
	case reflect.Slice:
		cfg := cfg.([]interface{})
		vs := reflect.MakeSlice(te, len(cfg), len(cfg))
		val := reflect.New(te.Elem())
		for k, vc := range cfg {
			gop.resolve(val, vc)

			vs.Index(k).Set(val.Elem())
		}
		ve.Set(vs)
	case reflect.Struct:
		gop.merge(ve, cfg.(map[string]interface{}))
	default:
		fv, err := format(ve, cfg)

		if err != nil {
			panic(err.Error())
		}
		ve.Set(reflect.ValueOf(fv))
	}
}

func format(vv reflect.Value, v interface{}) (interface{}, error) {
	tv := vv.Type()

	switch tv.Kind() {
	case reflect.String:
		return formatString(v)
	case reflect.Bool:
		return formatBool(v)

	case reflect.Int32, reflect.Int64:
		if tv.Name() == "Duration" {
			return formatDuration(v)
		}
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		i, err := formatInt64(v)
		if err != nil {
			return i, err
		}

		vi := reflect.ValueOf(i)
		return vi.Convert(tv).Interface(), nil

	case reflect.Float32, reflect.Float64:
		i, err := formatFloat64(v)
		if err != nil {
			return i, err
		}

		vi := reflect.ValueOf(i)
		return vi.Convert(tv).Interface(), nil
	case reflect.Interface:
		return reflect.ValueOf(v), nil
	default:
		return nil, fmt.Errorf("invalid option %s (%+v) value type %T", tv.Name(), tv.Kind(), v)
	}
}

func formatString(v interface{}) (string, error) {
	return fmt.Sprintf("%s", v), nil
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
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(v).Uint()), nil
	case string:
		return strconv.ParseInt(v.(string), 10, 64)
	}

	return 0, fmt.Errorf("invalid int64 value type %T", v)
}

func formatDuration(v interface{}) (time.Duration, error) {
	switch v.(type) {
	case time.Duration:
		return v.(time.Duration), nil
	case int, int8, int16, int32, int64, uint, uint16, uint32, uint64:
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
