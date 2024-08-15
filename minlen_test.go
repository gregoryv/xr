package xr

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// https://json-schema.org/draft-04/json-schema-validation#rfc.section.5.2.2

func TestPick_minLen(t *testing.T) {
	var x struct {
		Alias string `header:"alias" minLength:"5"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("alias", "ginger")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_minLenExceeded(t *testing.T) {
	var x struct {
		Alias string `header:"alias" minLength:"5"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("alias", "joe")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_minLenNotInteger(t *testing.T) {
	var x struct {
		Alias string `header:"alias" minLength:"jibberish"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("alias", "John Doe")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}
