package httpr

import (
	"bytes"
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

		if tag := field.Tag.Get("path"); tag != "" {
			val := r.PathValue(tag)
			err = set(obj, i, field, val)
		}

		if tag := field.Tag.Get("query"); tag != "" {
			val := query.Get(tag)
			err = set(obj, i, field, val)
		}
		if tag := field.Tag.Get("header"); tag != "" {
			val := r.Header.Get(tag)
			err = set(obj, i, field, val)
		}
	}
	return err
}

func set(obj reflect.Value, i int, field reflect.StructField, val string) error {
	if val == "" {
		return nil
	}

	elm := obj.Elem()
	// private fields cannot be set using reflect
	privateField := isPrivateField(elm.Type(), i)
	var setName string
	if privateField {
		setName = "Set" + capitalizeFirstLetter(field.Name)
	} else {
		setName = "Set" + field.Name
	}

	if fn := obj.MethodByName(setName); fn.IsValid() {
		_ = fn.Call([]reflect.Value{reflect.ValueOf(val)})
		return nil
	}

	if privateField {
		msg := fmt.Sprintf(
			"private field %s, missing %s", field.Name, setName,
		)
		panic(msg)
	}

	kind := field.Type.Kind()

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
func isPrivateField(t reflect.Type, i int) bool {
	field := t.Field(i)
	return field.PkgPath != ""
}
func capitalizeFirstLetter(s string) string {
	b := []byte(s)
	b[0] = bytes.ToUpper([]byte{b[0]})[0]

	return string(b)
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
