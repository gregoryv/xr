package httpr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExampleUnmarshal() {

	// handler on server side
	h := func(w http.ResponseWriter, r *http.Request) {
		var x PersonCreate
		_ = Decode(&x, r)
		fmt.Println("id:", x.Id)
		fmt.Println("name:", x.Name)
		fmt.Println("group:", x.Group)
		fmt.Println("copy:", x.Copy)
		fmt.Println("flag:", x.Flag)
		fmt.Println("token:", x.Token)
	}

	w := httptest.NewRecorder()
	data := `{"Name":"John Doe"}`
	body := strings.NewReader(data)
	r := httptest.NewRequest("POST", "/person/123?group=aliens&copies=10&flag=true", body)
	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", "Bearer ...token...")

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
}

type PersonCreate struct {
	Id    string `path:"id"`
	Name  string `json:"name"`
	Group string `query:"group"`
	Copy  int    `query:"copies"`
	Flag  bool   `query:"flag"`

	Token string `header:"authorization"`
	Auth  string `header:"authorization"`
}

func (p *PersonCreate) SetToken(v string) {
	if strings.HasPrefix(v, "Bearer ") {
		p.Token = v[7:]
		return
	}
	p.Token = v
}
