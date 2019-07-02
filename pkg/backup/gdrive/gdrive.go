package gdrive

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/config"
	gdrive "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/message"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/glower/file-watcher/notification"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
)

// Storage ...
type Storage struct {
	ctx context.Context

	name                  string // storage name
	globalConfigPath      string
	MessageCh             chan message.Message
	eventCh               chan notification.Event
	fileStorageProgressCh chan types.BackupProgress
	storagePath           string
	root                  *drive.File
	tokenFile             string
	credentialsFile       string
	client                *http.Client
	service               *drive.Service
}

const storageName = "storage.gdrive"

func init() {
	backup.Register(storageName, &Storage{})
}

// Setup gdrive storage
func (s *Storage) Setup(m *backup.StorageManager) (bool, error) {
	gdriveConfig := gdrive.GoogleDriveConfig()
	if gdriveConfig.Active {
		s.ctx = m.Ctx
		s.eventCh = make(chan notification.Event)
		s.fileStorageProgressCh = m.FileBackupProgressCh
		s.MessageCh = m.MessageCh

		s.globalConfigPath = config.GetConfigPath()
		s.credentialsFile = gdriveConfig.CredentialsFile
		s.tokenFile = gdriveConfig.TokenFile
		defaultPath := gdriveConfig.Path
		if defaultPath == "" {
			defaultPath = backup.DefultFolderName()
		}
		s.storagePath = defaultPath
		s.name = storageName
		credPath := filepath.Join(s.globalConfigPath, "credentials.json")
		b, err := ioutil.ReadFile(credPath)
		if err != nil {
			return false, fmt.Errorf("unable to read credentials file [%s]: %v", credPath, err)
		}

		config, err := google.ConfigFromJSON(b, drive.DriveScope)
		if err != nil {
			return false, fmt.Errorf("unable to parse client secret file to config: %v", err)
		}
		client, err := s.getClient(config)
		if err != nil {
			return false, fmt.Errorf("unable to create new client: %v", err)
		}

		// TODO: drive.New is deprecated, use something like this:
		// ctx := context.Background()
		// srv, err := drive.NewService(s.ctx, option.WithAPIKey("xbc"))
		srv, err := drive.New(client)
		if err != nil {
			return false, fmt.Errorf("unable to retrieve gDrive client: %v", err)
		}
		s.client = client
		s.service = srv

		s.root, err = s.CreateFolder(s.storagePath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Store ...
func (s *Storage) Store(event *notification.Event) error {
	to := remotePath(event.AbsolutePath, event.RelativePath)
	return s.store(event.AbsolutePath, to, event.MimeType)
}

// gdrive.store(): C:\Users\Brown\MyFiles\pixiv\71738080_p0_master1200.jpg > MyFiles\pixiv
func (s *Storage) store(file, toPath, mimeType string) error {
	sleepRandom()
	// log.Printf("[DEBUG] gdrive.store(): send [%s] -> [%s]\n", file, filepath.Join(s.storagePath, toPath))

	fromFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("cannot open file  [%s]: %v", file, err)
	}
	defer fromFile.Close()
	lastFolder, err := s.GetOrCreateAllFolders(toPath)
	if err != nil {
		return err
	}
	// if lastFolder == nil {
	// 	return fmt.Errorf("foler was not found or created")
	// }
	// fmt.Printf("[DEBUG] Create of update file %s in folder %s\n", filepath.Base(file), lastFolder.Name)
	_, err = s.CreateOrUpdateFile(fromFile, filepath.Base(file), mimeType, lastFolder.Id)
	if err != nil {
		return err
	}

	// log.Printf("[OK] gdrive.store(): %s, %s, %s DONE\n", gFile.Name, gFile.Id, gFile.MimeType)
	return nil
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
