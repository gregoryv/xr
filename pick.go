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
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// NewPicker returns a picker with no content-type decoders.
func NewPicker() *Picker {
	p := Picker{
		registry: make(map[string]func(io.Reader) Decoder),
		setters:  make(map[string]setfn),
		kindSetters: map[reflect.Kind]setfn{
			reflect.String: setStringField,

			reflect.Bool: setBoolField,

			reflect.Int:   setIntField,
			reflect.Int8:  setInt8Field,
			reflect.Int16: setInt16Field,
			reflect.Int32: setInt32Field,
			reflect.Int64: setInt64Field,

			reflect.Uint8:  setUint8Field,
			reflect.Uint16: setUint16Field,
			reflect.Uint32: setUint32Field,
			reflect.Uint64: setUint64Field,

			reflect.Float32: setFloat32Field,
			reflect.Float64: setFloat64Field,

			reflect.Complex64:  setComplex64Field,
			reflect.Complex128: setComplex128,
		},
	}
	return &p
}

type Picker struct {
	registry    map[string]func(io.Reader) Decoder
	setters     map[string]setfn
	kindSetters map[reflect.Kind]setfn
}

// Register body decoder based on content-type string.
func (p *Picker) Register(contentType string, fn func(io.Reader) Decoder) {
	p.registry[contentType] = fn
}

// UseSetter typ should be "package.Type"
func (p *Picker) UseSetter(typ string, fn setfn) {
	if _, found := p.setters[typ]; found {
		panic(fmt.Sprintf("UseSetter(%q): already exists", typ))
	}
	p.setters[typ] = fn
}

// Pick the given request into any struct type. Panics if dst is not a pointer.
func (p *Picker) Pick(dst any, r *http.Request) error {
	if t := reflect.TypeOf(dst); t.Kind() != reflect.Ptr {
		panic("Pick(dst, r): dst must be a pointer")
	}

	// decide for input format
	if err := p.decodeBody(dst, r); err != nil {
		return err
	}

	return p.pickFields(dst, r)
}

func (p *Picker) pickFields(dst any, r *http.Request) error {
	obj := reflect.ValueOf(dst)
	for i := 0; i < obj.Elem().NumField(); i++ {
		field := obj.Elem().Type().Field(i)
		tag := field.Tag

		val, source, err := readValue(r, tag)
		if errors.Is(err, errTagNotFound) {
			continue
		}

		if !field.IsExported() {
			panic(fmt.Sprintf("%v: private", field.Name))
		}
		if err := p.set(obj, i, val); err != nil {
			return &PickError{
				Dest:   obj.Elem().Type().Field(i).Name,
				Source: source,
				Cause:  err,
			}
		}
	}
	return nil
}

func (p *Picker) decodeBody(dst any, r *http.Request) error {
	switch r.Method {
	case "GET", "HEAD", "DELETE":
		// cannot have a body for decoding
		return nil

	default:
		ct := r.Header.Get("content-type")
		return p.newDecoder(ct, r.Body).Decode(dst)
	}
}

func (p *Picker) newDecoder(v string, r io.Reader) Decoder {
	if d, found := p.registry[v]; found {
		return d(r)
	}
	return noop
}

func readValue(r *http.Request, tag reflect.StructTag) (string, string, error) {
	for source, fn := range valueReaders {
		if v := tag.Get(source); v != "" {
			return fn(r, v), fmt.Sprintf("%s[%s]", source, v), nil
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

type (
	valueReader func(*http.Request, string) string
	setfn       func(field reflect.Value, v string) error
)

func (p *Picker) set(obj reflect.Value, i int, val string) error {
	if val == "" {
		return nil
	}
	field := obj.Elem().Type().Field(i)

	// find by type here
	fn, found := p.setters[field.Type.String()]
	if found {
		return fn(obj.Elem().Field(i), val)
	}

	kind := field.Type.Kind()
	fn, found = p.kindSetters[kind]
	if !found {
		return fmt.Errorf("set %v: unsupported", kind)
	}
	return fn(obj.Elem().Field(i), val)
}

func setBoolField(field reflect.Value, val string) error {
	value, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	field.SetBool(value)
	return nil
}

func setIntField(field reflect.Value, val string) error {
	value, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	field.SetInt(value)
	return nil
}

func setInt8Field(field reflect.Value, val string) error {
	value, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		return err
	}
	field.SetInt(value)
	return nil
}

func setInt16Field(field reflect.Value, val string) error {
	value, err := strconv.ParseInt(val, 10, 16)
	if err != nil {
		return err
	}
	field.SetInt(value)
	return nil
}

func setInt32Field(field reflect.Value, val string) error {
	value, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return err
	}
	field.SetInt(value)
	return nil
}

func setInt64Field(field reflect.Value, val string) error {
	value, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}
	field.SetInt(value)
	return nil
}

func setUint8Field(field reflect.Value, val string) error {
	value, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		return err
	}
	field.SetUint(value)
	return nil
}

func setUint16Field(field reflect.Value, val string) error {
	value, err := strconv.ParseUint(val, 10, 16)
	if err != nil {
		return err
	}
	field.SetUint(value)
	return nil
}

func setUint32Field(field reflect.Value, val string) error {
	value, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return err
	}
	field.SetUint(value)
	return nil
}

func setUint64Field(field reflect.Value, val string) error {
	value, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return err
	}
	field.SetUint(value)
	return nil
}

func setFloat32Field(field reflect.Value, val string) error {
	value, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return err
	}
	field.SetFloat(value)
	return nil
}

func setFloat64Field(field reflect.Value, val string) error {
	value, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	field.SetFloat(value)
	return nil
}

func setComplex64Field(field reflect.Value, val string) error {
	value, err := strconv.ParseComplex(val, 64)
	if err != nil {
		return err
	}
	field.SetComplex(value)
	return nil
}

func setStringField(field reflect.Value, val string) error {
	field.SetString(val)
	return nil
}

func setComplex128(field reflect.Value, val string) error {
	value, err := strconv.ParseComplex(val, 128)
	if err != nil {
		return err
	}
	field.SetComplex(value)
	return nil
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
	var cause string
	if e.Cause != nil {
		cause = strings.Replace(e.Cause.Error(), "strconv.", "", 1)
	}
	return fmt.Sprintf("pick %s from %s: %s", e.Dest, e.Source, cause)
}
