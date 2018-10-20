package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/handlers"
	"github.com/glower/bakku-app/pkg/watchers"
	"github.com/glower/bakku-app/pkg/watchers/watch"
	"github.com/r3labs/sse"

	// for auto import
	_ "github.com/glower/bakku-app/pkg/backup/storage/fake"
)

func init() {
	log.Println("init ...")
}

// TODO: try this out: https://github.com/antage/eventsource
func setupSSE() *sse.Server {
	events := sse.New()
	events.CreateStream("files")
	return events
}

func setupWatchers() []chan watch.FileChangeInfo {
	list := []chan watch.FileChangeInfo{}
	// TODO: check if the directrory is valid
	// TODO: check if `\` is at the end of the path,  it is importand!
	// directoriesToWatch := []string{`C:\Users\Brown\Downloads\`}
	directoriesToWatch := []string{`/home/igor/Downloads/`}
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
	sseServer := setupSSE()
	storageManager := storage.SetupManager(sseServer)
	startHTTPServer(sseServer, storageManager, fsWachers)

	// server will block here untill we got SIGTERM/kill
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	sseServer.RemoveStream("files")
	sseServer.Close()
	storage.Stop()
	log.Println("Shutdown the web server ...")
	// TODO: Shutdown is not working with open SSE connection, need to solve this first
	// srv.Shutdown(context.Background())
	log.Print("Done")
}

func startHTTPServer(sseServer *sse.Server, sManager *storage.Manager, fsWachers []chan watch.FileChangeInfo) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Port is not set, using default port 8080")
		port = "8080"
	}

	r := handlers.Resources{
		SSEServer: sseServer,
		FSChanges: fsWachers,
	}
	router := r.Router()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		for _, watcher := range fsWachers {
			for {
				select {
				case change := <-watcher:
					file := change.FileInfo
					info := fmt.Sprintf("Sync file [%s] it was %s\n", file.Name(), watch.ActionToString(change.Action))
					log.Print(info)
					// file the notification to the storage
					sManager.FileChangeNotificationChannel <- &storage.FileChangeNotification{
						File:   file,
						Action: change.Action,
					}
				}
			}
		}
	}()

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println("Failed to run server")
		}
	}()
	log.Print("The service is ready to listen and serve.")
	return srv
}
