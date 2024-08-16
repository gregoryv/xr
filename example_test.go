package xr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExamplePick_default() {
	// handler on server side
	h := func(w http.ResponseWriter, r *http.Request) {
		var x PersonCreate
		err := Pick(&x, r)
		fmt.Println("id:", x.Id)
		fmt.Println("name:", x.Name)
		fmt.Println("group:", x.Group)
		fmt.Println("copy:", x.Copy)
		fmt.Println("flag:", x.Flag)
		fmt.Println("pval:", x.PVal)
		fmt.Println("token:", x.token)
		fmt.Println("color:", x.Color)
		fmt.Println("width:", x.Width)
		fmt.Println(err)
	}

	w := httptest.NewRecorder()
	data := `{"Name":"John Doe", "Width": 100}`
	body := strings.NewReader(data)
	r := httptest.NewRequest("POST", "/person/123?group=aliens&copies=10&flag=true&pval=11.79", body)
	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", "Bearer ...token...")
	r.Header.Set("color", "yellow")

	mux := http.NewServeMux()
	mux.HandleFunc("/person/{id}", h)
	mux.ServeHTTP(w, r)
	// output:
	// id: 123
	// name: John Doe
	// group: aliens
	// copy: 10
	// flag: true
	// pval: 11.79
	// token: ...token...
	// color: yellow
	// width: 100
	// <nil>
}

type PersonCreate struct {
	Id    string  `path:"id"`
	Name  string  `json:"name" xml:"name"`
	Group string  `query:"group"`
	Copy  int     `query:"copies"`
	Flag  bool    `query:"flag"`
	PVal  float32 `query:"pval"`

	// private field requires method SetToken
	token string `header:"authorization"`

	Auth  string `header:"authorization"`
	Color string `header:"Color"`
	Width int    `json:"width" minimum:"200"`
}

// SetToken trims optional prefix "Bearer " from v.
func (p *PersonCreate) SetToken(v string) {
	if strings.HasPrefix(v, "Bearer ") {
		p.token = v[7:]
		return
	}
	p.token = v
}

func (p *PersonCreate) SetColor(v string) error {
	switch v {
	case "black":
		return fmt.Errorf("color unsupported: %s", v)
	default:
		p.Color = v
	}
	return nil
}
