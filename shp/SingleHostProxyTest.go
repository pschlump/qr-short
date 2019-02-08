package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		// Host:   "localhost:2004",
		Host: "192.168.0.157:2004",
	})
	http.ListenAndServe(":9002", proxy)
}

/*
---------------- Simple Server Example --------------------

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <port>", os.Args[0])
	}
	if _, err := strconv.Atoi(os.Args[1]); err != nil {
		log.Fatalf("Invalid port: %s (%s)\n", os.Args[1], err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		println("--->", os.Args[1], req.URL.String())
	})
	http.ListenAndServe(":"+os.Args[1], nil)
}

*/
