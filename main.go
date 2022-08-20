package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"simpletracker/simpletracker"
	"sync"
)

var tracker *simpletracker.Tracker

func handler(w http.ResponseWriter, r *http.Request) {
	return
}

func handlerTracker(w http.ResponseWriter, r *http.Request) {
	ts, err := tracker.TrackerToS()
	if err != nil {
		w.Write([]byte("unable to convert tracker to string"))
		if err != nil {
			return
		}
		w.WriteHeader(500)
		return
	}
	w.Write([]byte(ts))
	w.WriteHeader(200)
	return
}

func main() {
	syncMap := &sync.Map{}

	r := mux.NewRouter()
	r.HandleFunc("/", handler).Methods("GET")
	r.HandleFunc("/simpletracker", handler).Methods("POST")
	r.HandleFunc("/simpletracker", handler).Methods("GET")
	r.HandleFunc("/examplePath/{whatever}", handler).Methods("GET")
	r.HandleFunc("/tracker", handlerTracker).Methods("GET")

	tracker = simpletracker.New(syncMap, r, nil, nil)
	err := tracker.PopulateRoutes()
	if err != nil {
		log.Fatal(err)
	}

	r.Use(tracker.SimpleTracker)
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}

}
