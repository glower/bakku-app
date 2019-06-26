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
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/storage"
	"github.com/glower/bakku-app/pkg/types"

	// autoimport
	_ "github.com/glower/bakku-app/pkg"
	"github.com/glower/bakku-app/pkg/config"
	"github.com/glower/bakku-app/pkg/handlers"

	"github.com/glower/file-watcher/watcher"
)

func init() {
	config.ReadDefaultConfig()
}

func main() {
	log.Println("Starting the service ...")

	ctx, cancel := context.WithCancel(context.Background())
	fileWatcher := watcher.Setup(ctx,
		&watcher.Options{
			IgnoreDirectoies: true,
			FileFilters:      []string{".crdownload", ".lock", ".snapshot", ".snapshot.lock"},
		})

	res := types.GlobalResources{
		BackupCompleteCh: make(chan types.BackupComplete),
		MessageCh:        make(chan message.Message),
		FileWatcher:      fileWatcher,
		Storage:          storage.New(config.GetStoragePath()),
	}

	snapShotManager := snapshot.Setup(ctx, res)

	// read from the configuration file a list of directories to watch
	dirs, _ := config.DirectoriesToWatch()
	for _, d := range dirs.DirsToWatch {
		if d.Active {
			go fileWatcher.StartWatching(d.Path)
			go snapShotManager.CreateOrUpdate(d.Path)
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	startHTTPServer(res)

	// eventBuffer := event.NewBuffer(ctx, res)
	// fmt.Println("event buffer is up and running ...")

	// backupStorageManager := backup.Setup(ctx, res, eventBuffer)
	// fmt.Println("backup storage manager is up and running ...")

	// sseServer := event.NewSSE(ctx, router, backupStorageManager.FileBackupProgressCh, res, eventBuffer)
	// fmt.Println("SSE server is up and running ...")

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
	// sseServer.StopSSE()
	backup.Stop()
	log.Println("Shutdown the web server ...")
	// TODO: Shutdown is not working with open SSE connection, need to solve this first
	// srv.Shutdown(context.Background())
}

func startHTTPServer(res types.GlobalResources) *mux.Router {
	port := os.Getenv("BAKKU_PORT")
	if port == "" {
		log.Println("Port is not set, using default port 8080")
		port = "8080"
	}

	r := handlers.Resources{
		FileWatcher: res.FileWatcher,
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
	log.Print("[OK] The service is ready to listen and serve!")
	return router
}

func initStorage() {

}
