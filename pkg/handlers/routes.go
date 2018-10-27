package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/r3labs/sse"
)

// Resources ...
type Resources struct {
	SSEServer *sse.Server
	// FSChanges []chan watch.FileChangeInfo
}

// Router register necessary routes and returns an instance of a router.
func (res *Resources) Router() *mux.Router {
	r := mux.NewRouter()
	r.Methods("GET").Path("/health").HandlerFunc(StatusOK)
	r.Methods("GET").Path("/ping").HandlerFunc(Ping)
	r.Methods("GET").Path("/events").HandlerFunc(res.SSEServer.HTTPHandler)

	return r
}

// StatusOK returns the status code 200
func StatusOK(res http.ResponseWriter, _ *http.Request) {
	res.WriteHeader(http.StatusOK)
}

// Ping returns pong
func Ping(res http.ResponseWriter, _ *http.Request) {
	res.WriteHeader(http.StatusOK)
	fmt.Fprint(res, "pong")
}
