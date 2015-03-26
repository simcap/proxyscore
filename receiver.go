package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/", home)
	http.ListenAndServe(":4673", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	e := json.NewEncoder(w)
	e.Encode(r)
}
