package xr

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecode_unknownTag(t *testing.T) {
	var x struct {
		Jib bool `jib:"first"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	if err := Decode(&x, r); err != nil {
		t.Error(err)
	}
}

func TestDecode_stopOnFirstError(t *testing.T) {
	var x struct {
		First  bool `header:"first"`
		Second bool `header:"second"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("first", "jibberish")
	r.Header.Set("second", "true")
	if err := Decode(&x, r); err == nil {
		t.Error("expect error")
	}
	if x.Second {
		t.Error("Second was set")
	}
}

func TestDecode_contentType(t *testing.T) {
	data := `{broken`
	body := strings.NewReader(data)
	r := httptest.NewRequest("GET", "/", body)
	r.Header.Set("content-type", "application/json")

	var x struct {
		Name string `json:"name"`
	}
	if err := Decode(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestDecode_unsupported(t *testing.T) {
	type complex struct {
		Name string
	}
	var x struct {
		C complex `header:"input"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("input", "not an int")
	if err := Decode(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestDecode_bool(t *testing.T) {
	var x struct {
		I bool `header:"flag"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("flag", "not an int")
	if err := Decode(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestDecode_atoi(t *testing.T) {
	var x struct {
		I int `header:"number"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "not an int")
	if err := Decode(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestDecode_missingSet(t *testing.T) {
	defer catchPanic(t)
	var x struct {
		// private, needs a SetToken
		token string `header:"authorization"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("authorization", "Bearer ...token...")
	_ = Decode(&x, r)
}

func TestDecode_noValue(t *testing.T) {
	var x struct {
		token string `header:"authorization"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	// missing header
	if err := Decode(&x, r); err != nil {
		t.Error(err)
	}
}

func catchPanic(t *testing.T) {
	if err := recover(); err == nil {
		t.Fatal("expect panic")
	}
}
