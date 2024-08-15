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
		setters:  make(map[string]setfn),
	}
}

type Picker struct {
	registry map[string]func(io.Reader) Decoder
	setters  map[string]setfn
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

// Pick the given request into any struct type.
func (p *Picker) Pick(dst any, r *http.Request) *PickError {
	if t := reflect.TypeOf(dst); t.Kind() != reflect.Ptr {
		panic("Pick(dst, r): dst must be a pointer")
	}

	// decide for input format
	switch r.Method {
	case "GET", "HEAD", "DELETE":
		// cannot have a body for decoding
	default:
		ct := r.Header.Get("content-type")
		dec := p.newDecoder(ct, r.Body)
		if err := dec.Decode(dst); err != nil {
			return &PickError{
				Dest:   fmt.Sprintf("%T", dst)[1:],
				Source: "body",
				Cause:  err,
			}
		}
	}

	obj := reflect.ValueOf(dst)
	for i := 0; i < obj.Elem().NumField(); i++ {

		tag := obj.Elem().Type().Field(i).Tag
		val, source, err := readValue(r, tag)
		if !errors.Is(err, errTagNotFound) {
			if err := p.set(obj, i, val); err != nil {
				return &PickError{
					Dest:   fmt.Sprintf("%v.%s", obj.Elem().Type(), obj.Elem().Type().Field(i).Name),
					Source: source,
					Cause:  err,
				}
			}
		}
		// once all fields are set, validate against field tags, eg. minLength
		if err := minLength(tag, val); err != nil {
			return &PickError{
				Dest:   fmt.Sprintf("%v.%s", obj.Elem().Type(), obj.Elem().Type().Field(i).Name),
				Source: source,
				Cause:  err,
			}
		}
		if err := maxLength(tag, val); err != nil {
			return &PickError{
				Dest:   fmt.Sprintf("%v.%s", obj.Elem().Type(), obj.Elem().Type().Field(i).Name),
				Source: source,
				Cause:  err,
			}
		}
		if err := minimumField(obj, i, val); err != nil {
			return &PickError{
				Dest:   fmt.Sprintf("%v.%s", obj.Elem().Type(), obj.Elem().Type().Field(i).Name),
				Source: source,
				Cause:  err,
			}
		}
		if err := maximumField(obj, i, val); err != nil {
			return &PickError{
				Dest:   fmt.Sprintf("%v.%s", obj.Elem().Type(), obj.Elem().Type().Field(i).Name),
				Source: source,
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

type (
	valueReader func(*http.Request, string) string
	setfn       func(field reflect.Value, v string) error
)

func (p *Picker) set(obj reflect.Value, i int, val string) error {
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

	// find by type here
	fn, found := p.setters[field.Type.String()]
	if found {
		return fn(obj.Elem().Field(i), val)
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

func minLength(tag reflect.StructTag, value string) error {
	in, found := tag.Lookup("minLength")
	if !found {
		return nil
	}
	min, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return err
	}
	if int64(len(value)) < min {
		return fmt.Errorf("minLength exceeded")
	}
	return nil
}

func maxLength(tag reflect.StructTag, value string) error {
	in, found := tag.Lookup("maxLength")
	if !found {
		return nil
	}
	max, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return err
	}
	if int64(len(value)) > max {
		return fmt.Errorf("maxLength exceeded")
	}
	return nil
}

func minimumField(obj reflect.Value, i int, in string) error {
	// get minimum value as float
	field := obj.Elem().Type().Field(i)
	min, found := field.Tag.Lookup("minimum")
	if !found {
		return nil
	}
	minVal, err := strconv.ParseFloat(min, 32)
	if err != nil {
		return err
	}
	// get the already set value
	var value float64
	kind := field.Type.Kind()
	switch kind {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(obj.Elem().Field(i).Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(obj.Elem().Field(i).Int())
	case reflect.Float32, reflect.Float64:
		value = float64(obj.Elem().Field(i).Float())
	}
	if value < minVal {
		return fmt.Errorf("%v, minimum %v exceeded", value, minVal)
	}
	return nil
}

func maximumField(obj reflect.Value, i int, in string) error {
	// get maximum value as float
	field := obj.Elem().Type().Field(i)
	max, found := field.Tag.Lookup("maximum")
	if !found {
		return nil
	}
	maxVal, err := strconv.ParseFloat(max, 32)
	if err != nil {
		return err
	}
	// get the already set value
	var value float64
	kind := field.Type.Kind()
	switch kind {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value = float64(obj.Elem().Field(i).Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value = float64(obj.Elem().Field(i).Int())
	case reflect.Float32, reflect.Float64:
		value = float64(obj.Elem().Field(i).Float())
	}
	if value > maxVal {
		return fmt.Errorf("maximum exceeded")
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
	return fmt.Sprintf("%s: %s", e.Dest, e.Cause.Error())
}
