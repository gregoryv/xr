package xr

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExampleRegister_xml() {
	// register decoder for content-type header
	Register(
		"application/xml",
		func(r io.Reader) Decoder {
			return xml.NewDecoder(r)
		},
	)
	h := func(w http.ResponseWriter, r *http.Request) {
		var x struct {
			Name  string `xml:"name"`
			Width int    `xml:"width"`
		}
		_ = Pick(&x, r)
		fmt.Println("name:", x.Name)
		fmt.Println("width:", x.Width)
	}

	w := httptest.NewRecorder()
	data := `<person>
<name>John Doe</name>
<width>100</width>
</person>`
	body := strings.NewReader(data)
	r := httptest.NewRequest("POST", "/person", body)
	r.Header.Set("content-type", "application/xml")

	h(w, r)
	// output:
	// name: John Doe
	// width: 100
}
