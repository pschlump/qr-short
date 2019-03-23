package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pschlump/godebug"
)

// // GetVar returns a variable by name from GET or POST data.
// func GetVar(name string, www http.ResponseWriter, req *http.Request) (found bool, value string) {
// 	method := req.Method
// 	if method == "POST" {
// 		if str := req.PostFormValue(name); str != "" {
// 			value = str
// 			found = true
// 		}
// 	} else if method == "GET" {
// 		if str := req.URL.Query().Get(name); str != "" {
// 			value = str
// 			found = true
// 		}
// 	} else {
// 		www.WriteHeader(418) // Ha Ha - I Am A Tea Pot
// 	}
// 	return
// }

// GetVar returns a variable by name from GET or POST data.
func GetVar(name string, www http.ResponseWriter, req *http.Request) (found bool, value string) {
	method := req.Method
	if db_flag["GetVal"] {
		fmt.Printf("GetVar name=%s req.Method %s AT:%s\n", name, method, godebug.LF())
	}
	if method == "POST" || method == "PUT" {
		if str := req.PostFormValue(name); str != "" { // xyzzy - actually have to check if exists
			value = str
			found = true
		}
	} else if method == "GET" || method == "DELETE" {
		if db_flag["GetVal"] {
			fmt.Printf("AT:%s\n", godebug.LF())
		}
		qq := req.URL.Query()
		strArr, ok := qq[name]
		if db_flag["GetVal"] {
			fmt.Printf("AT:%s strArr = %s ok = %v\n", godebug.LF(), godebug.SVar(strArr), ok)
		}
		if ok {
			if db_flag["GetVal"] {
				fmt.Printf("AT:%s\n", godebug.LF())
			}
			if len(strArr) > 0 {
				if db_flag["GetVal"] {
					fmt.Printf("AT:%s\n", godebug.LF())
				}
				value = strArr[0]
				found = true
			} else {
				if db_flag["GetVal"] {
					fmt.Printf("AT:%s\n", godebug.LF())
				}
				fmt.Fprintf(os.Stderr, "Multiple values for [%s]\n", name)
				found = false
			}
		}
	} else {
		www.WriteHeader(418) // Ha Ha - I Am A Tea Pot
	}
	return
}
