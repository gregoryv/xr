package xr

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// See
// https://json-schema.org/draft-04/json-schema-validation#rfc.section.5.1.2

func TestPick_minimumUint8(t *testing.T) {
	var x struct {
		I uint8 `header:"number" minimum:"5"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "5")
	if err := Pick(&x, r); err != nil {
		t.Error(err)
	}
}

func TestPick_minimumExceeded(t *testing.T) {
	cases := map[string]struct {
		X   any
		Val string
	}{
		"uint8": {
			new(struct {
				I uint8 `header:"number" minimum:"5"`
			}),
			"2",
		},
		"uint16": {
			new(struct {
				I uint16 `header:"number" minimum:"5"`
			}),
			"2",
		},
		"uint32": {
			new(struct {
				I uint32 `header:"number" minimum:"5"`
			}),
			"2",
		},
		"uint64": {
			new(struct {
				I uint64 `header:"number" minimum:"5"`
			}),
			"2",
		},
		"int": {
			new(struct {
				I int `header:"number" minimum:"-5"`
			}),
			"-10",
		},
		"int8": {
			new(struct {
				I int8 `header:"number" minimum:"-5"`
			}),
			"-10",
		},
		"int16": {
			new(struct {
				I int16 `header:"number" minimum:"-5"`
			}),
			"-10",
		},
		"int32": {
			new(struct {
				I int32 `header:"number" minimum:"-5"`
			}),
			"-10",
		},
		"int64": {
			new(struct {
				I int64 `header:"number" minimum:"-5"`
			}),
			"-10",
		},
		"float32": {
			new(struct {
				I float32 `header:"number" minimum:"-5.14"`
			}),
			"-5.15",
		},
		"float64": {
			new(struct {
				I float64 `header:"number" minimum:"-5.14"`
			}),
			"-5.15",
		},
		"float64e": {
			new(struct {
				I float64 `header:"number" minimum:"1.5e3"`
			}),
			"1.2e2",
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", http.NoBody)
			r.Header.Set("number", c.Val)
			if err := Pick(c.X, r); err == nil {
				t.Error("expect error")
			}
		})
	}
}

func TestPick_minimumNotFloat(t *testing.T) {
	var x struct {
		I uint8 `header:"number" minimum:"jibberish"`
	}
	r := httptest.NewRequest("GET", "/", http.NoBody)
	r.Header.Set("number", "2")
	if err := Pick(&x, r); err == nil {
		t.Error("expect error")
	}
}
