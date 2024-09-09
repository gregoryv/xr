package xr

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExamplePick_default() {
	// handler on server side
	h := func(w http.ResponseWriter, r *http.Request) {
		var x PersonCreate
		if err := Pick(&x, r); err != nil {
			log.Fatal(err)
		}
		fmt.Print(x)
	}

	w := httptest.NewRecorder()
	data := `{"Name":"John Doe", "Width": 100}`
	body := strings.NewReader(data)
	u := "/person/123?group=aliens&copies=10&flag=true&pval=11.79"
	r := httptest.NewRequest("POST", u, body)
	r.Header.Set("content-type", "application/json; charset=utf-8")
	r.Header.Set("authorization", "Bearer ...token...")
	r.Header.Set("color", "yellow")

	mux := http.NewServeMux()
	mux.HandleFunc("/person/{id}", h)
	mux.ServeHTTP(w, r)
	// output:
	// {123 John Doe aliens 10 true 11.79 Bearer ...token... yellow 100 }
}

type PersonCreate struct {
	Id    string  `path:"id"`
	Name  string  `json:"name" xml:"name"`
	Group string  `query:"group"`
	Copy  int     `query:"copies"`
	Flag  bool    `query:"flag"`
	PVal  float32 `query:"pval"`

	Auth  string `header:"authorization"`
	Color string `header:"Color"`
	Width int    `json:"width"`

	// private fields are ignored
	token string
}
