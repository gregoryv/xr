[gregoryv/xr](https://pkg.go.dev/github.com/gregoryv/xr) - Pick values from a http.Request

**Pick the meat of a bone**

Similar to decoding the body it supports reading values from
the path, query and forms. See [examples](https://pkg.go.dev/github.com/gregoryv/xr#pkg-examples).

The package supports a subset of tag names as defined by
[swaggest/openapi-go](https://github.com/swaggest/openapi-go).

In addition, set methods are used for additional format checking

    type Account struct {
	    token string `header:"authorization"`
    }
	func (x *Account) SetToken(v string) error {
	    ...
	}

Setters can return any number of values as long as the last one is an
error it will be returned.

