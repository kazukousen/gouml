package main

import (
	"net/http"
)

func main() {
	http.Handle("/gouml", goumlHandler())

	http.ListenAndServe(":8080", nil)
}

func goumlHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
	}
	return http.HandlerFunc(fn)
}
