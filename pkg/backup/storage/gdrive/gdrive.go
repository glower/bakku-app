package gdrive

import (
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	backupstorage "github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/config"
	gdrive "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/file-watcher/notification"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// Storage ...
type Storage struct {
	name                  string // storage name
	globalConfigPath      string
	eventCh               chan notification.Event
	fileStorageProgressCh chan types.BackupProgress
	ctx                   context.Context
	storagePath           string
	root                  *drive.File
	tokenFile             string
	credentialsFile       string
	client                *http.Client
	service               *drive.Service
}

const storageName = "gdrive"

func init() {
	backupstorage.Register(storageName, &Storage{})
}

// Setup gdrive storage
func (s *Storage) Setup(fileStorageProgressCh chan types.BackupProgress) bool {
	gdriveConfig := gdrive.GoogleDriveConfig()
	if gdriveConfig.Active {
		s.eventCh = make(chan notification.Event)
		s.fileStorageProgressCh = fileStorageProgressCh

		s.globalConfigPath = config.GetConfigPath()
		s.credentialsFile = gdriveConfig.CredentialsFile
		s.tokenFile = gdriveConfig.TokenFile
		defaultPath := gdriveConfig.Path
		if defaultPath == "" {
			defaultPath = backupstorage.DefultFolderName()
		}
		s.storagePath = defaultPath
		s.name = storageName
		credPath := filepath.Join(s.globalConfigPath, "credentials.json")
		b, err := ioutil.ReadFile(credPath)
		if err != nil {
			log.Fatalf("[ERROR] gdrive.Setup(): Unable to read credentials file [%s]: %v", credPath, err)
			return false
		}

		config, err := google.ConfigFromJSON(b, drive.DriveScope)
		if err != nil {
			log.Fatalf("[ERROR] gdrive.Setup(): Unable to parse client secret file to config: %v", err)
		}
		client := s.getClient(config)
		srv, err := drive.New(client)
		if err != nil {
			log.Fatalf("[ERROR] gdrive.Setup(): Unable to retrieve Drive client: %v", err)
		}
		s.client = client
		s.service = srv

		s.root = s.CreateFolder(s.storagePath)
		return true
	}
	return false
}

// Store ...
func (s *Storage) Store(event *notification.Event) {
	to := remotePath(event.AbsolutePath, event.RelativePath)
	s.store(event.AbsolutePath, to, "image/jpeg")
}

// gdrive.store(): C:\Users\Brown\MyFiles\pixiv\71738080_p0_master1200.jpg > MyFiles\pixiv
func (s *Storage) store(file, toPath, mimeType string) {
	sleepRandom()
	log.Printf("gdrive.store(): send [%s] -> [gdrive://%s]\n", file, toPath)

	from, err := os.Open(file)
	if err != nil {
		log.Fatalf("[ERROR] gdrive.store(): Cannot open file  [%s]: %v\n", file, err)
		return
	}
	defer from.Close()
	lastFolder := s.GetOrCreateAllFolders(toPath) // TODO: errors?

	f := &drive.File{
		Name:     filepath.Base(file),
		MimeType: mimeType,
		Parents:  []string{lastFolder.Id},
	}

	// DefaultUploadChunkSize = 8 * 1024 * 1024
	chunkSize := googleapi.ChunkSize(5 * 1024 * 1024)
	contentType := googleapi.ContentType(mimeType)

	// TODO: wrap this with createOrUpdateFile
	res, err := s.service.Files.Create(f).Media(from, chunkSize, contentType).Do()

	if err != nil {
		log.Fatalf("[ERROR] gdrive.store(): %v", err)
	}
	log.Printf("gdrive.store(): %s, %s, %s DONE\n", res.Name, res.Id, res.MimeType)
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

func sleepRandom() {
	r := 500000 + rand.Intn(2000000)
	time.Sleep(time.Duration(r) * time.Microsecond)
}
