package xr

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// See
//
//	https://json-schema.org/draft-04/json-schema-validation#rfc.section.5.2.1

func TestPick_maxLen(t *testing.T) {
	var x struct {
		Alias string `header:"alias" maxLength:"5"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("alias", "joe")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_maxLenExceeded(t *testing.T) {
	var x struct {
		Alias string `header:"alias" maxLength:"5"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("alias", "John Doe")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}

func TestPick_maxLenNotInteger(t *testing.T) {
	var x struct {
		Alias string `header:"alias" maxLength:"jibberish"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("alias", "John Doe")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}
