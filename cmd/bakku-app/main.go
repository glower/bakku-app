package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"

	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/handlers"

	"github.com/glower/file-watcher/notification"
	"github.com/glower/file-watcher/watcher"

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

func processProgressCallback(ctx context.Context, fileBackupProgressChannel chan types.BackupProgress, sseServer *sse.Server) {
	log.Printf("\n")
	for {
		select {
		case <-ctx.Done():
			return
		case progress := <-fileBackupProgressChannel:
			// TODO: don't report progress on backup of snapshot file
			// get this name from config/package
			if strings.Contains(progress.FileName, ".snapshot") {
				continue
			}
			log.Printf("ProcessProgressCallback(): [%s] [%s]\t%.2f%%\n", progress.StorageName, progress.FileName, progress.Percent)
			progressJSON, _ := json.Marshal(progress)
			// file fotification for the frontend client over the SSE
			sseServer.Publish("files", &sse.Event{
				Data: []byte(progressJSON),
			})
		}
	}
}

func processErrors(ctx context.Context, errorCh chan notification.Error) {
	for {
		select {
		case <-ctx.Done():
			return
		case err := <-errorCh:
			if err.Level == "ERROR" || err.Level == "CRITICAL" {
				log.Printf("[%s] %v\n", err.Level, err.Message)
				fmt.Println("-----------------------------")
				fmt.Printf("%v\n", err.Stack)
				fmt.Println("-----------------------------")
			}
		}
	}
}

func main() {
	log.Println("Starting the service ...")
	ctx, cancel := context.WithCancel(context.Background())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	sseServer := setupSSE()

	// read from the configuration file a list of directories to watch
	dirs := config.DirectoriesToWatch()

	eventCh, errorCh := watcher.Setup(
		ctx,
		dirs,
		[]notification.ActionType{},
		[]string{".crdownload", ".lock", ".snapshot", ".snapshot.lock"}, // TODO: move me to some config
		&watcher.Options{IgnoreDirectoies: true})

	go processErrors(ctx, errorCh)

	backupStorageManager := backup.Setup(ctx, eventCh)
	snapshot.Setup(ctx, eventCh, backupStorageManager.FileBackupCompleteCh)

	go processProgressCallback(ctx, backupStorageManager.FileBackupProgressCh, sseServer)
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
	port := os.Getenv("BAKKU_PORT")
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
			log.Printf("[ERROR] Failed to run server: %v", err)
		}
	}()
	log.Print("The service is ready to listen and serve.")
	return srv
}
