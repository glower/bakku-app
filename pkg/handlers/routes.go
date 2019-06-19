package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/file-watcher/watcher"

	"github.com/gorilla/mux"
)

// Resources ...
type Resources struct {
	FileWatcher *watcher.Watch
	// TODO: need here
	// 1. file watcher object
	// 2. snapshot manager here
	// OR some sort of config manager
}

// Router register necessary routes and returns an instance of a router.
func (res *Resources) Router() *mux.Router {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").HandlerFunc(Index)
	r.Methods("GET").Path("/health").HandlerFunc(StatusOK)
	r.Methods("GET").Path("/ping").HandlerFunc(Ping)

	r.Methods("GET").Path("/api/config").HandlerFunc(Config)
	r.Methods("PSOT").Path("/api/config").HandlerFunc(res.UpdateConfig)

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

// Config returns local configuration
func Config(w http.ResponseWriter, r *http.Request) {
	// TODO: return complete configuration
	conf, err := config.DirectoriesToWatch()
	if err != nil {
		ServerError(w, err.Error())
		return
	}
	json, err := conf.ToJSON()
	if err != nil {
		ServerError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(json))
}

func (res *Resources) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. Add/Remove directories to file watcher
	// 	  watcher.Add(dir)/watcher.Remove(dir)
	// 2. Scan new dir:
	//    snapshot.Add(dir)
	// 3. Update the config
	// OR call some config manager and let the manager manage all this

	conf, err := config.FromJSON(r.Body)
	if err != nil {
		ServerError(w, err.Error())
		return
	}
	fmt.Printf("api.UpdateConfig(): %v\n", conf)
	w.WriteHeader(201)
}

func ServerError(w http.ResponseWriter, m string) {
	w.WriteHeader(500)
	w.Write([]byte(fmt.Sprintf(`{"error", "%s"}`, m)))
}
