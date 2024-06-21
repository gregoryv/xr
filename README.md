[gregoryv/xr](https://pkg.dev.go/github.com/gregoryv/xr) - Pick values from a http.Request

This package simplifies picking values into a struct from a request.

Early in a http.Request processing you often read values from
different parts of the request. Usually those values end up in one
struct of sorts. That is where package xr comes in. Making the picking
of values from URLs, headers and body easier.

It uses reflection to check for field tags, such as; **path, query, header**

To support checking values in this same process the xr.Picker looks
for matching Set**FieldName** methods on a struct which may or may not
return an error.

    type Account struct {
	    token string `header:"authorization"`
    }
	func (x *Account) SetToken(v string) error {
	    ...
	}


This also works as a way to pick values into private fields.

