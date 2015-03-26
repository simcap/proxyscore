package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/score", score)
	http.ListenAndServe(":4673", nil)
}

var results store = store{}

type store struct {
	requests map[string](*http.Request)
	sync.Mutex
}

func (s *store) add(id string, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	if s.requests == nil {
		s.requests = make(map[string](*http.Request))
	}

	s.requests[id] = r
}

func home(w http.ResponseWriter, r *http.Request) {
	e := json.NewEncoder(w)
	e.Encode(r)
}

func submit(w http.ResponseWriter, r *http.Request) {
	identifier := r.URL.Query().Get("u")
	if len(identifier) != 8 {
		http.Error(w, "invalid identifier", http.StatusBadRequest)
	}

	results.add(identifier, r)
}

func score(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	identifier := r.URL.Query().Get("u")
	if req, ok := results.requests[identifier]; ok {
		e := json.NewEncoder(w)
		e.Encode(req)
	} else {
		http.Error(w, "identifier not found", http.StatusNotFound)
	}
}
