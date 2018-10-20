package handlers

import (
	"fmt"
	"net/http"

	"github.com/glower/file-change-notification/watch"
	"github.com/gorilla/mux"
	"github.com/r3labs/sse"
)

// Resources ...
type Resources struct {
	SSEServer *sse.Server
	FSChanges []chan watch.FileChangeInfo
}

// Router register necessary routes and returns an instance of a router.
func (res *Resources) Router() *mux.Router {
	r := mux.NewRouter()
	r.Methods("GET").Path("/health").HandlerFunc(StatusOK)
	r.Methods("GET").Path("/ping").HandlerFunc(Ping)
	r.Methods("GET").Path("/events").HandlerFunc(res.SSEServer.HTTPHandler)

	// go func() {
	// 	for _, watcher := range res.FSChanges {
	// 		for {
	// 			select {
	// 			case change := <-watcher:
	// 				file := change.FileInfo
	// 				info := fmt.Sprintf("Sync file [%s] it was %s\n", file.Name(), watch.ActionToString(change.Action))
	// 				log.Println(info)
	// 				res.SSEServer.Publish("files", &sse.Event{
	// 					Data: []byte(info),
	// 				})
	// 			}
	// 		}
	// 	}
	// }()

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
