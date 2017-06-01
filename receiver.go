package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
)

var portFlag = flag.String("p", ":4673", "Service port")

func main() {
	flag.Parse()

	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*portFlag, nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	d := struct {
		Headers    http.Header
		RemoteAddr string
	}{
		Headers:    r.Header,
		RemoteAddr: r.RemoteAddr,
	}

	json.NewEncoder(w).Encode(d)
}
