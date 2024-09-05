package xr

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func BenchmarkPick(b *testing.B) {
	// incoming request
	var buf bytes.Buffer
	buf.WriteString(`{"Name":"John Doe", Width: 231}`)

	u := "/person/123?group=aliens&copies=10&flag=true"
	r := httptest.NewRequest("POST", u, &buf)
	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", "Bearer ...token...")

	var x PersonCreate
	for i := 0; i < b.N; i++ {
		_ = Pick(&x, r)
	}
}
