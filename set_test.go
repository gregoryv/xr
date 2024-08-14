package xr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPicker_UseSetter_duplicate(t *testing.T) {
	defer catchPanic(t)
	p := NewPicker()
	p.UseSetter("xr.Color", SetColorField)
	p.UseSetter("xr.Color", SetColorField)
}

func TestPicker_typeX(t *testing.T) {
	// Configure picker to use our set func for the specific type.
	// Using global UseSetter in this test for coverage.
	UseSetter("xr.Color", SetColorField)

	var x struct {
		I Color `header:"color"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("color", "yellow")

	// ok
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
	if x.I != Yellow {
		t.Error("got", x.I, "exp", Yellow)
	}
}

func TestPicker_typeX_fail(t *testing.T) {
	p := NewPicker()
	p.UseSetter("xr.Color", SetColorField)

	var x struct {
		I Color `header:"color"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("color", "neon")

	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

type Color int

const (
	Black = iota
	Red
	Yellow
)

func SetColorField(field reflect.Value, v string) error {
	color, err := ParseColor(v)
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(color))
	return nil
}

func ParseColor(v string) (Color, error) {
	switch v {
	case "yellow":
		return Yellow, nil
	case "red":
		return Red, nil
	case "black":
		return Black, nil
	default:
		return Black, fmt.Errorf("unknown color: %v", v)
	}
}
