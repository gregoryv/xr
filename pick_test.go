package xr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPick_pickErrors(t *testing.T) {
	var x Car
	r := httptest.NewRequest("GET", "/?model=ford", http.NoBody)
	err := Pick(&x, r)
	if err.Source != "query[model]" {
		t.Error("invalid source", err)
	}
	if err.Dest != "xr.Car.model" {
		t.Error("invalid destination", err)
	}

	{ // decoding
		data := `{"sold": "some-incorrect-value"}`
		body := strings.NewReader(data)
		r := httptest.NewRequest("POST", "/", body)
		r.Header.Set("content-type", "application/json")

		err := Pick(&x, r)
		if err.Source != "body" {
			t.Error("bad source", err.Source)
		}
	}
}

type Car struct {
	model string `query:"model"`

	Sold bool `json:"sold"`
}

func (c *Car) SetModel(v string) error {
	if v != "audi" {
		return fmt.Errorf("unknown model %q", v)
	}
	c.model = v
	return nil
}

func TestPick_nonPointer(t *testing.T) {
	defer catchPanic(t)
	var x struct {
		Jib bool `query:"jib"`
	}
	r := httptest.NewRequest("GET", "/?jib=true", http.NoBody)
	Pick(x, r)
	if !x.Jib {
		t.Fail()
	}
}

func TestPick_unknownTag(t *testing.T) {
	var x struct {
		Jib bool `jib:"first"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_stopOnFirstError(t *testing.T) {
	var x struct {
		First  bool `header:"first"`
		Second bool `header:"second"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("first", "jibberish")
	r.Header.Set("second", "true")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	if x.Second {
		t.Error("Second was set")
	}
}

func TestPick_contentType(t *testing.T) {
	data := `{broken`
	body := strings.NewReader(data)
	r := httptest.NewRequest("GET", "/", body)
	r.Header.Set("content-type", "application/json")

	var x struct {
		Name string `json:"name"`
	}
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_unsupported(t *testing.T) {
	type thing struct {
		Name string
	}
	var x struct {
		I thing `header:"input"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("input", "not an int")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_bool(t *testing.T) {
	var x struct {
		I bool `header:"flag"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("flag", "not an int")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_int(t *testing.T) {
	var x struct {
		I int `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "jibberish")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_int8(t *testing.T) {
	var x struct {
		I int8 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "-129")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "-128")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_int16(t *testing.T) {
	var x struct {
		I int16 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "-32769")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "-32768")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_int32(t *testing.T) {
	var x struct {
		I int32 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "-2147483649")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "-2147483648")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_int64(t *testing.T) {
	var x struct {
		I int64 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "-9223372036854775809")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "-9223372036854775808")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_uint8(t *testing.T) {
	var x struct {
		I uint8 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "256")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "255")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_uint16(t *testing.T) {
	var x struct {
		I uint16 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "65536")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "65535")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_uint32(t *testing.T) {
	var x struct {
		I uint32 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "4294967296")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "4294967295")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_uint64(t *testing.T) {
	var x struct {
		I uint64 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "18446744073709551616")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "18446744073709551615")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_complex64(t *testing.T) {
	var x struct {
		I complex64 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	// test value taken from strconv/atoc_test.go
	r.Header.Set("number", "2e308+2e308i")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error", x.I)
	}
	// ok case
	r.Header.Set("number", "-1.175494351e-38")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_complex128(t *testing.T) {
	var x struct {
		I complex128 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "2e308+2e308i")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error", x.I)
	}
	// ok case
	r.Header.Set("number", "-1.175494351e-38")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_float32(t *testing.T) {
	var x struct {
		I float32 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "not a float32")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_float64(t *testing.T) {
	var x struct {
		I float64 `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "not a float64")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
	// ok case
	r.Header.Set("number", "-123.99")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_atoi(t *testing.T) {
	var x struct {
		I int `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "not an int")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_missingSet(t *testing.T) {
	defer catchPanic(t)
	var x struct {
		// private, needs a SetToken
		token string `header:"authorization"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("authorization", "Bearer ...token...")
	_ = Pick(&x, r)
}

func TestPick_noValue(t *testing.T) {
	var x struct {
		token string `header:"authorization"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	// missing header
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func catchPanic(t *testing.T) {
	if err := recover(); err == nil {
		t.Fatal("expect panic")
	}
}
