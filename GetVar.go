package main

import (
	"net/http"
)

// GetVar returns a variable by name from GET or POST data.
func GetVar(name string, www http.ResponseWriter, req *http.Request) (found bool, value string) {
	method := req.Method
	if method == "POST" {
		if str := req.PostFormValue(name); str != "" {
			value = str
			found = true
		}
	} else if method == "GET" {
		if str := req.URL.Query().Get(name); str != "" {
			value = str
			found = true
		}
	} else {
		www.WriteHeader(418) // Ha Ha - I Am A Tea Pot
	}
	return
}
