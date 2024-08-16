package xr

import (
	"encoding/json"
	"io"
	"net/http"
)

func init() {
	p := NewPicker()
	p.Register("application/json",
		func(r io.Reader) Decoder {
			return json.NewDecoder(r)
		},
	)
	PickerDefault = p
}

// Pick using [PickerDefault]
func Pick(dst any, r *http.Request) error {
	return PickerDefault.Pick(dst, r)
}

// Register using [PickerDefault]
func Register(contentType string, fn func(io.Reader) Decoder) {
	PickerDefault.Register(contentType, fn)
}

// UseSetter using [PickerDefault]
func UseSetter(typ string, fn setfn) {
	PickerDefault.UseSetter(typ, fn)
}

// PickerDefault has a predefined content-type decoder for
// application/json.
var PickerDefault *Picker
