package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/gorilla/mux"
)

// Resources ...
type Resources struct {
	// SSEServer *sse.Server
	// FSChanges []chan watch.FileChangeInfo
}

// Router register necessary routes and returns an instance of a router.
func (res *Resources) Router() *mux.Router {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").HandlerFunc(Index)
	r.Methods("GET").Path("/health").HandlerFunc(StatusOK)
	r.Methods("GET").Path("/ping").HandlerFunc(Ping)

	r.Methods("GET").Path("/api/config").HandlerFunc(Config)

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

// Index returns index page with test UI
func Index(res http.ResponseWriter, _ *http.Request) {
	body, _ := ioutil.ReadFile("ui/index.html")
	fmt.Fprintf(res, "%s", body)
}

// type ConfigJSON struct {
// 	Directories []string ``
// }

// Config returns local configuration
func Config(w http.ResponseWriter, r *http.Request) {
	// TODO: return complete configuration
	dirs := config.DirectoriesToWatch()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dirs)
}
