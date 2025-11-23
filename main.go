package main

import (
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Server hello =)")
}

func inspect(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func favicon(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "resources/favicon/ms-icon-310x310.png")
}

func main() {
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/inspect", inspect)
	http.HandleFunc("/favicon.ico", favicon)

	http.ListenAndServe(":8090", nil)
}
