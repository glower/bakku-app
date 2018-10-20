package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/handlers"
	"github.com/glower/bakku-app/pkg/watchers"
	"github.com/glower/file-change-notification/watch"
	"github.com/r3labs/sse"
)

func init() {
	log.Println("init ...")
}

func setupSSE() *sse.Server {
	events := sse.New()
	events.CreateStream("files")
	return events
}

func setupWatchers() []chan watch.FileChangeInfo {
	list := []chan watch.FileChangeInfo{}
	// TODO: check if the directrory is valid
	// TODO: check if `\` is at the end of the path,  it is importand!
	directoriesToWatch := []string{`C:\Users\Brown\Downloads\`}
	for _, dir := range directoriesToWatch {
		watcher := watchers.WatchDirectoryForChanges(dir)
		list = append(list, watcher)
	}
	return list
}

func main() {
	log.Println("Starting the service ...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	fsWachers := setupWatchers()
	events := setupSSE()
	srv := startHTTPServer(events, fsWachers)

	// setup storages for the backup
	storage.Run()

	// server will block here untill we got SIGTERM/kill
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	srv.Shutdown(context.Background())
	events.Close()
	storage.Stop()
	log.Print("Done")
}

func startHTTPServer(events *sse.Server, fsWachers []chan watch.FileChangeInfo) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Port is not set, using default port 8080")
		port = "8080"
	}

	r := handlers.Resources{
		SSEServer: events,
		FSChanges: fsWachers,
	}
	router := r.Router()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Panicf("Httpserver: ListenAndServe() error: %s\n", err)
		}
	}()
	log.Print("The service is ready to listen and serve.")
	return srv
}
