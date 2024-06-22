package xr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func ExamplePick_form() {
	// handler on server side
	h := func(w http.ResponseWriter, r *http.Request) {
		var x struct {
			Name string `form:"name"`
		}
		_ = Pick(&x, r)
		fmt.Println("name:", x.Name)
	}

	w := httptest.NewRecorder()
	form := make(url.Values)
	form.Set("name", "John Doe")
	data := form.Encode()
	body := strings.NewReader(data)
	r := httptest.NewRequest("POST", "/person", body)
	r.Header.Set("content-type", "application/x-www-form-urlencoded")
	h(w, r)

	// output:
	// name: John Doe
}
