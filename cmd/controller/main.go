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
	"github.com/spf13/viper"

	// for auto import
	_ "github.com/glower/bakku-app/pkg/backup/storage/fake"
	// _ "github.com/glower/bakku-app/pkg/backup/storage/local"
)

func init() {
	log.Println("init ...")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
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
	for _, dir := range viper.Get("watch").([]interface{}) {
		watcher := watchers.WatchDirectoryForChanges(dir.(string))
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

	go handleFileChangedRequests(sManager.FileChangeNotificationChannel, fsWachers)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println("Failed to run server")
		}
	}()
	log.Print("The service is ready to listen and serve.")
	return srv
}

// TODO: https://go101.org/article/channel-use-cases.html (Rate Limiting)
func handleFileChangedRequests(fileChangeNotificationChannel chan *storage.FileChangeNotification, fsWachers []chan watch.FileChangeInfo) {
	for _, watcher := range fsWachers {
		for {
			select {
			case change := <-watcher:
				log.Printf("main.handleFileChangedRequests(): event for [%s]\n", change.FileInfo.Name())
				// file the notification to the storage
				fileChangeNotificationChannel <- &storage.FileChangeNotification{
					File:   change.FileInfo,
					Path:   change.FilePath,
					Action: change.Action,
				}
			}
		}
	}
}
