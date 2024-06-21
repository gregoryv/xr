// Package xr provides a http.Request decode func.
//
// Pick first tries to decode the body based on the request
// content-type header. E.g. "application/json" will use json.Decoder.
//
// If successfull, field tags are used to decode the rest.  For each
// field tag of a struct the value is read and set.  If there is a
// method named Set{FIELD_TAG}, it is used, otherwise field is set
// directly using reflection.
package xr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

// Pick the given request into any struct type.
func Pick(dst any, r *http.Request) error {
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
	typ := elm.Type()

	for i := 0; i < elm.NumField(); i++ {
		field := typ.Field(i)

		val, err := readValue(r, field.Tag)
		if errors.Is(err, errTagNotFound) {
			continue
		}
		if err := set(obj, i, field, val); err != nil {
			return err
		}
	}
	return nil
}

func readValue(r *http.Request, tag reflect.StructTag) (string, error) {
	for t, fn := range valueReaders {
		if v := tag.Get(t); v != "" {
			return fn(r, v), nil
		}
	}
	return "", errTagNotFound
}

var errTagNotFound = errors.New("tag not found")

// valueReaders map how field tags are read from a given request
var valueReaders = map[string]valueReader{
	"path": func(r *http.Request, name string) string {
		return r.PathValue(name)
	},
	"query": func(r *http.Request, name string) string {
		return r.URL.Query().Get(name)
	},
	"header": func(r *http.Request, name string) string {
		return r.Header.Get(name)
	},
}

type valueReader func(*http.Request, string) string

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
		result := fn.Call([]reflect.Value{reflect.ValueOf(val)})
		i := len(result)
		if i > 0 && !result[i-1].IsNil() {
			return result[i-1].Interface().(error)
		}
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
	if d, found := registry[v]; found {
		return d(r)
	}
	return noop
}

func Register(contentType string, fn func(io.Reader) Decoder) {
	registry[contentType] = fn
}

// registry maps a content-type string to a decoder
var registry = map[string]func(io.Reader) Decoder{
	"application/json": func(r io.Reader) Decoder {
		return json.NewDecoder(r)
	},
}

var noop = decoderFunc(func(_ any) error { return nil })

type decoderFunc func(v any) error

func (fn decoderFunc) Decode(v any) error {
	return fn(v)
}

type Decoder interface {
	Decode(v any) error
}
