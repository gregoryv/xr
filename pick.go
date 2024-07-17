// Package xr provides means to pick values from a http.Request
//
// Pick first tries to decode the body based on the content-type
// header. E.g. "application/json" will use json.Decoder.
//
// If successfull, field tags are used to decode the rest.  For each
// field tag of a struct the value is read and set.  If there is a
// method named Set{FIELD_TAG}, it is used, otherwise field is set
// directly using reflection.
package xr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

// NewPicker returns a picker with no content-type decoders.
func NewPicker() *Picker {
	return &Picker{
		registry: make(map[string]func(io.Reader) Decoder),
	}
}

type Picker struct {
	registry map[string]func(io.Reader) Decoder
}

// Register body decoder based on content-type string.
func (p *Picker) Register(contentType string, fn func(io.Reader) Decoder) {
	p.registry[contentType] = fn
}

// Pick the given request into any struct type.
func (p *Picker) Pick(dst any, r *http.Request) *PickError {
	if t := reflect.TypeOf(dst); t.Kind() != reflect.Ptr {
		panic("Pick(dst, r): dst must be a pointer")
	}

	// decide for input format
	ct := r.Header.Get("content-type")
	dec := p.newDecoder(ct, r.Body)
	if err := dec.Decode(dst); err != nil {
		return &PickError{
			Dest:   fmt.Sprintf("%T", dst)[1:],
			Source: "body",
			Cause:  err,
		}
	}

	obj := reflect.ValueOf(dst)
	for i := 0; i < obj.Elem().NumField(); i++ {

		val, tag, err := readValue(r, obj.Elem().Type().Field(i).Tag)
		if errors.Is(err, errTagNotFound) {
			continue
		}
		if err := set(obj, i, val); err != nil {
			return &PickError{
				Dest:   fmt.Sprintf("%v.%s", obj.Elem().Type(), obj.Elem().Type().Field(i).Name),
				Source: tag,
				Cause:  err,
			}
		}
	}
	return nil
}

func (p *Picker) newDecoder(v string, r io.Reader) Decoder {
	if d, found := p.registry[v]; found {
		return d(r)
	}
	return noop
}

func readValue(r *http.Request, tag reflect.StructTag) (string, string, error) {
	for t, fn := range valueReaders {
		if v := tag.Get(t); v != "" {
			return fn(r, v), fmt.Sprintf("%s[%s]", t, v), nil
		}
	}
	return "", "", errTagNotFound
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
	"form": func(r *http.Request, name string) string {
		return r.FormValue(name)
	},
}

type valueReader func(*http.Request, string) string

func set(obj reflect.Value, i int, val string) error {
	if val == "" {
		return nil
	}
	field := obj.Elem().Type().Field(i)
	// private fields cannot be set using reflect
	isPrivateField := field.PkgPath != ""
	var setMethod string
	if isPrivateField {
		setMethod = "Set" + capitalizeFirstLetter(field.Name)
	} else {
		setMethod = "Set" + field.Name
	}

	if fn := obj.MethodByName(setMethod); fn.IsValid() {
		result := fn.Call([]reflect.Value{reflect.ValueOf(val)})
		// return error from setMethod, if any
		i := len(result)
		if i > 0 && !result[i-1].IsNil() {
			return result[i-1].Interface().(error)
		}
		return nil
	}

	if isPrivateField {
		msg := fmt.Sprintf(
			"private field %s, missing %s", field.Name, setMethod,
		)
		panic(msg)
	}

	kind := field.Type.Kind()
	switch kind {
	case reflect.Bool:
		value, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetBool(value)

	case reflect.Int:
		value, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetInt(value)

	case reflect.Int8:
		value, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetInt(value)

	case reflect.Int16:
		value, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetInt(value)

	case reflect.Int32:
		value, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetInt(value)

	case reflect.Int64:
		value, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetInt(value)

	case reflect.Uint8:
		value, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetUint(value)

	case reflect.Uint16:
		value, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetUint(value)

	case reflect.Uint32:
		value, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetUint(value)

	case reflect.Uint64:
		value, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetUint(value)

	case reflect.Float32:
		value, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetFloat(value)

	case reflect.Float64:
		value, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetFloat(value)

	case reflect.Complex64:
		value, err := strconv.ParseComplex(val, 64)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetComplex(value)

	case reflect.Complex128:
		value, err := strconv.ParseComplex(val, 128)
		if err != nil {
			return err
		}
		obj.Elem().Field(i).SetComplex(value)

	case reflect.String:
		obj.Elem().Field(i).SetString(val)

		// add more types when needed
	default:
		return fmt.Errorf("set %v: unsupported", kind)
	}
	return nil
}

func capitalizeFirstLetter(s string) string {
	b := []byte(s)
	b[0] = bytes.ToUpper([]byte{b[0]})[0]
	return string(b)
}

var noop = decoderFunc(func(_ any) error { return nil })

type decoderFunc func(v any) error

func (fn decoderFunc) Decode(v any) error {
	return fn(v)
}

type Decoder interface {
	Decode(v any) error
}

type PickError struct {
	// package.type.field
	Dest string

	// (path|query|header|form)[NAME] or body, e.g. header[correlationId]
	Source string

	// parsing or set error
	Cause error
}

func (e *PickError) Error() string {
	return fmt.Sprintf("pick %s from %s: %s", e.Dest, e.Source, e.Cause.Error())
}
