package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/snapshot"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/handlers"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/bakku-app/pkg/watchers"
	"github.com/r3labs/sse"

	// for auto import
	_ "github.com/glower/bakku-app/pkg/backup/storage/fake"
	_ "github.com/glower/bakku-app/pkg/backup/storage/gdrive"
	_ "github.com/glower/bakku-app/pkg/backup/storage/local"
)

func init() {
	config.ReadDefaultConfig()
}

// TODO: try this out: https://github.com/antage/eventsource
func setupSSE() *sse.Server {
	events := sse.New()
	events.CreateStream("files")
	return events
}

func main() {
	log.Println("Starting the service ...")
	ctx, cancel := context.WithCancel(context.Background())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	sseServer := setupSSE()

	// each time a file is changed or created we will get a notification on this channel
	fileChangeNotificationChan := make(chan types.FileChangeNotification)

	watchers.SetupFSWatchers(fileChangeNotificationChan)
	backupStorageManager := backup.Setup(ctx, fileChangeNotificationChan)
	snapshot.Setup(fileChangeNotificationChan, backupStorageManager.FileBackupCompleteChannel)
	startHTTPServer(sseServer)

	// server will block here untill we got SIGTERM/kill
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	cancel()
	sseServer.RemoveStream("files")
	sseServer.Close()
	backup.Stop()
	log.Println("Shutdown the web server ...")
	// TODO: Shutdown is not working with open SSE connection, need to solve this first
	// srv.Shutdown(context.Background())
}

func startHTTPServer(sseServer *sse.Server) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Port is not set, using default port 8080")
		port = "8080"
	}

	r := handlers.Resources{
		SSEServer: sseServer,
	}
	router := r.Router()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Println("Failed to run server")
		}
	}()
	log.Print("The service is ready to listen and serve.")
	return srv
}
