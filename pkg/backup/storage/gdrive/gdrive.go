package gdrive

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

		credPath := filepath.Join(s.globalConfigPath, "credentials.json")
		b, err := ioutil.ReadFile(credPath)
		if err != nil {
			log.Fatalf("gdrive.Setup(): Unable to read credentials file [%s]: %v", credPath, err)
			return false
		}

		config, err := google.ConfigFromJSON(b, drive.DriveScope)
		if err != nil {
			log.Fatalf("gdrive.Setup(): Unable to parse client secret file to config: %v", err)
		}
		client := s.getClient(config)
		srv, err := drive.New(client)
		if err != nil {
			log.Fatalf("gdrive.Setup(): Unable to retrieve Drive client: %v", err)
		}
		s.client = client
		s.service = srv

		s.root = s.CreateFolder(s.storagePath)

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

// TODO: move me to storage/backup namespace!
func (s *Storage) handleFileChanges(fileChange *types.FileChangeNotification) {
	log.Printf("gdrive.handleFileChanges(): File [%#v] has been changed\n", fileChange)
	absolutePath := fileChange.AbsolutePath // /foo/bar/buz/alice.jpg
	relativePath := fileChange.RelativePath // buz/alice.jpg
	// directoryPath := fileChange.DirectoryPath // /foo/bar/

	from := absolutePath
	to := remotePath(absolutePath, relativePath)

	// localSnapshotPath := snapshot.StoragePath(directoryPath)
	// remoteSnapshotPath := snapshot.StoragePath(s.storagePath)

	// don't backup file if it is in progress
	if ok := storage.BackupStarted(absolutePath, storageName); ok {
		s.store(from, to)
		storage.BackupFinished(absolutePath, storageName)
		// snapshot.UpdateEntry(directoryPath, relativePath)
		// s.SyncSnapshot(localSnapshotPath, remoteSnapshotPath)
	}
}

func (s *Storage) store(fromPath, toPath string) {
	from, err := os.Open(fromPath)
	if err != nil {
		log.Printf("[ERROR] gdrive.store(): Cannot open file  [%s]: %v\n", fromPath, err)
		return
	}
	defer from.Close()
	lastFolder := s.CreateAllFolders(toPath) // TODO: errors?

	f := &drive.File{
		Name:     filepath.Base(fromPath),
		MimeType: "image/jpeg",
		Parents:  []string{lastFolder.Id},
	}
	res, err := s.service.Files.Create(f).Media(from).Do()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("%s, %s, %s\n", res.Name, res.Id, res.MimeType)

	// fromStrats, _ := from.Stat()
}

func remotePath(absolutePath, relativePath string) string {
	absolutePathArr := strings.Split(absolutePath, string(os.PathSeparator))
	relativePathArr := strings.Split(relativePath, string(os.PathSeparator))
	result := []string{}
	for i, part := range absolutePathArr {
		if part == string(relativePathArr[0]) {
			result = absolutePathArr[i-1 : len(absolutePathArr)-1]
		}
	}
	return strings.Join(result, string(os.PathSeparator))
}
