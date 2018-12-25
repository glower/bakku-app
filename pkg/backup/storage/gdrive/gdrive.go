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

	"github.com/glower/bakku-app/pkg/backup"
	"github.com/glower/bakku-app/pkg/backup/storage"
	"github.com/glower/bakku-app/pkg/config"
	gdrive "github.com/glower/bakku-app/pkg/config/storage"
	"github.com/glower/bakku-app/pkg/snapshot"
	"github.com/glower/bakku-app/pkg/types"
	"github.com/otiai10/copy"
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
	tokenFile                     string
	credentialsFile               string
	client                        *http.Client
	service                       *drive.Service
}

const storageName = "gdrive"

func init() {
	storage.Register(storageName, &Storage{})
}

// Setup gdrive storage
func (s *Storage) Setup(fileStorageProgressCannel chan *storage.Progress) bool {
	gdriveConfig := gdrive.GoogleDriveConfig()
	if gdriveConfig.Active {
		s.fileChangeNotificationChannel = make(chan *types.FileChangeNotification)
		s.fileStorageProgressCannel = fileStorageProgressCannel

		s.globalConfigPath = config.GetConfigPath()
		s.credentialsFile = gdriveConfig.CredentialsFile
		s.tokenFile = gdriveConfig.TokenFile
		defaultPath := gdriveConfig.Path
		if defaultPath == "" {
			defaultPath = backup.DefultFolderName()
		}
		s.storagePath = defaultPath

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

// SyncLocalFilesToBackup XXXX
func (s *Storage) SyncLocalFilesToBackup() {
	log.Println("gdrive.SyncLocalFilesToBackup(): START")
	dirs := config.DirectoriesToWatch()
	for _, path := range dirs {
		log.Printf("gdrive.SyncLocalFilesToBackup(): %s\n", path)

		remoteSnapshot := filepath.Join(s.storagePath, filepath.Base(path), snapshot.FileName(path))
		localTMPPath := filepath.Join(os.TempDir(), backup.DefultFolderName(), storageName, filepath.Base(path))
		localTMPFile := filepath.Join(localTMPPath, snapshot.FileName(path))

		log.Printf("gdrive.SyncLocalFilesToBackup(): copy snapshot for [%s] from [%s] to [%s]\n",
			path, remoteSnapshot, localTMPFile)

		if err := copy.Copy(remoteSnapshot, localTMPFile); err != nil {
			log.Printf("[ERROR] gdrive.SyncLocalFilesToBackup(): can't copy snapshot for [%s]: %v\n", path, err)
			return
		}

		s.syncFiles(localTMPPath, path)
	}
}

// syncFiles XXXX
func (s *Storage) syncFiles(remoteSnapshotPath, localSnapshotPath string) {
	log.Printf("gdrive.syncFiles(): from remote: [%s] to local [%s]\n", remoteSnapshotPath, localSnapshotPath)
	files, err := snapshot.Diff(remoteSnapshotPath, localSnapshotPath)
	if err != nil {
		log.Printf("[ERROR] gdrive.syncFiles(): %v\n", err)
		return
	}
	for _, file := range *files {
		s.fileChangeNotificationChannel <- &file
	}
}

// Store ...
func (s *Storage) Store(fileChange *types.FileChangeNotification) {
	to := remotePath(fileChange.AbsolutePath, fileChange.RelativePath)
	s.store(fileChange.AbsolutePath, to)
}

// SyncSnapshot syncs the snapshot dir to the storage
func (s *Storage) SyncSnapshot(fileChange *types.FileChangeNotification) {
	absolutePath := fileChange.AbsolutePath   // /foo/bar/buz/alice.jpg
	relativePath := fileChange.RelativePath   // buz/alice.jpg
	directoryPath := fileChange.DirectoryPath // /foo/bar/

	from := snapshot.FilePath(directoryPath)
	to := filepath.Join(remotePath(absolutePath, relativePath), snapshot.FileName(directoryPath))
	s.store(from, to)
	log.Printf("gdrive.SyncSnapshot(): sync snapshot from [%s] to [gdrive:%s]\n", from, to)

}

func (s *Storage) store(fromPath, toPath string) {
	sleepRandom()
	log.Printf("gdrive.store(): %s > %s\n", fromPath, toPath)
	from, err := os.Open(fromPath)
	if err != nil {
		log.Fatalf("[ERROR] gdrive.store(): Cannot open file  [%s]: %v\n", fromPath, err)
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
