package httpr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

func Decode(dst any, r *http.Request) error {
	// decide for input format
	dec := newDecoder(
		r.Header.Get("content-type"),
		r.Body,
	)
	if err := dec.Decode(dst); err != nil {
		return err
	}

	obj := reflect.ValueOf(dst)
	elm := obj.Elem()
	query := r.URL.Query()

	typ := elm.Type()

	var err error
	for i := 0; i < elm.NumField(); i++ {
		if err != nil {
			break
		}
		field := typ.Field(i)
		kind := field.Type.Kind()

		if tag := field.Tag.Get("path"); tag != "" {
			val := r.PathValue(tag)
			err = set(obj, i, kind, field.Name, val)
		}

		if tag := field.Tag.Get("query"); tag != "" {
			val := query.Get(tag)
			err = set(obj, i, kind, field.Name, val)
		}
		if tag := field.Tag.Get("header"); tag != "" {
			val := r.Header.Get(tag)
			err = set(obj, i, kind, field.Name, val)
		}
	}
	return err
}

func set(obj reflect.Value, i int, kind reflect.Kind, fieldName, val string) error {
	if val == "" {
		return nil
	}

	if fn := obj.MethodByName("Set" + fieldName); fn.IsValid() {
		_ = fn.Call([]reflect.Value{reflect.ValueOf(val)})
		return nil
	}

	elm := obj.Elem()
	switch kind {
	case reflect.Int:
		value, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		elm.Field(i).SetInt(int64(value))

	case reflect.String:
		elm.Field(i).SetString(val)

	case reflect.Bool:
		value, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		elm.Field(i).SetBool(value)

		// add more types when needed
	default:
		return fmt.Errorf("Unsupported VType %v", kind)
	}
	return nil
}

func newDecoder(v string, r io.Reader) Decoder {
	switch v {
	case "application/json":
		return json.NewDecoder(r)

	default:
		return noop
	}
}

var noop = DecoderFunc(func(_ any) error { return nil })

type Decoder interface {
	Decode(v any) error
}

type DecoderFunc func(v any) error

func (fn DecoderFunc) Decode(v any) error {
	return fn(v)
}
