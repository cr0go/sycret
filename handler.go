package main

import (
	"net/http"
)

func handler() {
	handle := http.HandlerFunc(handleRequest)
	http.Handle("/", handle)
	http.ListenAndServe(":8000", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	setQuery(r)
	initial(w)
}

func setQuery(r *http.Request) {
	keys := r.URL.Query()
	rqu = rquery{}

	for k, v := range keys {
		if k == "URLTemplate" {
			rqu.templateURL = v[0]
		}
		if k == "RecordID" {
			rqu.RecordID = v[0]
		}
	}
}
