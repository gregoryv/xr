package httpr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

func ExampleDecode() {
	// handler on server side
	h := func(w http.ResponseWriter, r *http.Request) {
		var x PersonCreate
		err := Decode(&x, r)
		fmt.Println("id:", x.Id)
		fmt.Println("name:", x.Name)
		fmt.Println("group:", x.Group)
		fmt.Println("copy:", x.Copy)
		fmt.Println("flag:", x.Flag)
		fmt.Println("token:", x.token)
		fmt.Println("color:", x.Color)
		fmt.Println("width:", x.Width)
		fmt.Println(err)
	}

	w := httptest.NewRecorder()
	data := `{"Name":"John Doe"}`
	body := strings.NewReader(data)
	r := httptest.NewRequest("POST", "/person/123?group=aliens&copies=10&flag=true", body)
	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", "Bearer ...token...")
	r.Header.Set("color", "yellow")
	r.Header.Set("width", "100cm")

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
	// width: 0
	// SetWidth("100cm"): strconv.Atoi: parsing "100cm": invalid syntax
}

type PersonCreate struct {
	Id    string `path:"id"`
	Name  string `json:"name" xml:"name"`
	Group string `query:"group"`
	Copy  int    `query:"copies"`
	Flag  bool   `query:"flag"`

	// private field requires method SetToken
	token string `header:"authorization"`

	Auth  string `header:"authorization"`
	Color string `header:"Color"`
	Width int    `header:"Width"`
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

// SetWidth returns old width, or 0 and error
func (p *PersonCreate) SetWidth(v string) (int, error) {
	val, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("SetWidth(%q): %w", v, err)
	}
	old := p.Width
	p.Width = val
	return old, nil
}
