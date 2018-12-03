package gdrive

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/config"
	gdrive "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
)

// Storage ...
type Storage struct {
	name                          string // storage name
	globalConfigPath              string
	fileChangeNotificationChannel chan *types.FileChangeNotification
	fileStorageProgressCannel     chan *storage.Progress
	ctx                           context.Context
	storagePath                   string
	root                          *drive.File
	clientID                      string
	clientSecret                  string
	client                        *http.Client
	service                       *drive.Service
}

const storageName = "gdrive"

func init() {
	storage.Register(storageName, &Storage{})
}

// SyncSnapshot syncs the snapshot dir to the storage
func (s *Storage) SyncSnapshot(from, to string) {}

// Setup gdrive storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	gdriveConfig := gdrive.GoogleDriveConfig()
	if gdriveConfig.Active {
		log.Println("storage.gdrive.Setup()")

		s.globalConfigPath = config.GetConfigPath()
		s.clientID = gdriveConfig.ClientID
		s.clientSecret = gdriveConfig.ClientSecret
		defaultPath := gdriveConfig.Path
		if defaultPath == "" {
			defaultPath = backup.DefultFolderName()
		}
		s.storagePath = defaultPath
		s.root = s.CreateFolder(s.storagePath)

		credPath := filepath.Join(s.globalConfigPath, "credentials.json")
		b, err := ioutil.ReadFile(credPath)
		if err != nil {
			log.Fatalf("gdrive.Setup(): Unable to read credentials file [%s]: %v", credPath, err)
			return false
		}

		config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := s.getClient(config)
		srv, err := drive.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve Drive client: %v", err)
		}
		s.client = client
		s.service = srv

		// Test
		r, err := srv.Files.List().PageSize(10).Fields("nextPageToken, files(id, name)").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve files: %v", err)
		}
		fmt.Println("Files:")
		if len(r.Files) == 0 {
			fmt.Println("No files found.")
		} else {
			for _, i := range r.Files {
				fmt.Printf("%s (%s)\n", i.Name, i.Id)
			}
		}

		return true
	}
	return false
}

// SyncLocalFilesToBackup ...
func (s *Storage) SyncLocalFilesToBackup() {}

// FileChangeNotification returns channel for notifications
func (s *Storage) FileChangeNotification() chan *types.FileChangeNotification {
	return s.fileChangeNotificationChannel
}

// Start local storage
func (s *Storage) Start(ctx context.Context) error {
	log.Println("storage.local.Start()")
	s.ctx = ctx
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case fileChange := <-s.fileChangeNotificationChannel:
				go s.handleFileChanges(fileChange)
			}
		}
	}()
	return nil
}

// handleFileChanges ...
func (s *Storage) handleFileChanges(fileChange *types.FileChangeNotification) {

}
