package httpr

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExampleDecode_xml() {
	// register decoders for content-type headers if needed
	// application/json is registered out of the box
	Register(
		"application/xml",
		func(r io.Reader) Decoder {
			return xml.NewDecoder(r)
		},
	)
	h := func(w http.ResponseWriter, r *http.Request) {
		var x PersonCreate
		_ = Decode(&x, r)
		fmt.Println("id:", x.Id)
		fmt.Println("name:", x.Name)
		fmt.Println("group:", x.Group)
		fmt.Println("copy:", x.Copy)
		fmt.Println("flag:", x.Flag)
		fmt.Println("token:", x.token)
		fmt.Println("color:", x.Color)
		fmt.Println("width:", x.Width)
	}

	w := httptest.NewRecorder()
	data := `<person><name>John Doe</name></person>`
	body := strings.NewReader(data)
	r := httptest.NewRequest("POST", "/person/123?group=aliens&copies=10&flag=true", body)
	r.Header.Set("content-type", "application/xml")
	r.Header.Set("authorization", "Bearer ...token...")
	r.Header.Set("color", "yellow")
	r.Header.Set("width", "100")

	mux := http.NewServeMux()
	mux.HandleFunc("/person/{id}", h)
	mux.ServeHTTP(w, r)
	// output:
	// id: 123
	// name: John Doe
	// group: aliens
	// copy: 10
	// flag: true
	// token: ...token...
	// color: yellow
	// width: 100
}
