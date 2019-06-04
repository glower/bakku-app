package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/event"
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"

	// autoimport
	_ "github.com/glower/bakku-app/pkg"
	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/handlers"

	"github.com/glower/file-watcher/notification"
	"github.com/glower/file-watcher/watcher"
)

func init() {
	config.ReadDefaultConfig()
}

func main() {
	log.Println("Starting the service ...")

	ctx, cancel := context.WithCancel(context.Background())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// read from the configuration file a list of directories to watch
	dirs := config.DirectoriesToWatch()

	eventCh, errorCh := watcher.Setup(
		ctx,
		dirs,
		[]notification.ActionType{},
		[]string{".crdownload", ".lock", ".snapshot", ".snapshot.lock"}, // TODO: move me to some config
		&watcher.Options{IgnoreDirectoies: true})

	messageCh := make(chan message.Message)
	backupDoneCh := make(chan types.FileBackupComplete)

	eventBuffer := event.New(ctx, eventCh, backupDoneCh)

	backupStorageManager := backup.Setup(ctx, eventBuffer.EvenOutCh, messageCh, backupDoneCh)

	snapshot.Setup(ctx, dirs, eventCh, messageCh, backupDoneCh)

	router := startHTTPServer()

	sseServer := event.NewSSE(ctx, router, backupStorageManager.FileBackupProgressCh, errorCh, messageCh)

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
	sseServer.StopSSE()
	backup.Stop()
	log.Println("Shutdown the web server ...")
	// TODO: Shutdown is not working with open SSE connection, need to solve this first
	// srv.Shutdown(context.Background())
}

func startHTTPServer() *mux.Router {
	port := os.Getenv("BAKKU_PORT")
	if port == "" {
		log.Println("Port is not set, using default port 8080")
		port = "8080"
	}

	r := handlers.Resources{}
	router := r.Router()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Printf("[ERROR] Failed to run server: %v", err)
		}
	}()
	log.Print("[OK] The service is ready to listen and serve.")
	return router
}
