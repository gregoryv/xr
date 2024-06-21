package httpr

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func catchPanic(t *testing.T) {
	if err := recover(); err == nil {
		t.Fatal("expect panic")
	}
}
