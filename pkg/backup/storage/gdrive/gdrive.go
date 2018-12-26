package gdrive

import (
	"context"
	"fmt"
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
	fileStorageProgressCannel     chan *backup.Progress
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
func (s *Storage) Setup(fileStorageProgressCannel chan *backup.Progress) bool {
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
	s.store(fileChange.AbsolutePath, to, "image/jpeg")
}

// SyncSnapshot syncs the snapshot dir to the storage
// FileChangeNotification{
// 		Action: 5,
// 		Machine: "DESKTOP-T75G9FM",
// 		Name: "69633416_p0.png",
// 		AbsolutePath: "C:\Users\Brown\MyFiles\pixiv\69633416_p0.png",
// 		RelativePath: "pixiv\69633416_p0.png",
// 		DirectoryPath: "C:\Users\Brown\MyFiles",
// 		WatchDirectoryName: "MyFiles",
// 		Size: 2567216,
// 		Timestamp: ...
// }
func (s *Storage) SyncSnapshot(fileChange *types.FileChangeNotification) {
	fmt.Printf("\n\n%#v\n\n", fileChange)
	// absolutePath := fileChange.AbsolutePath
	// relativePath := fileChange.RelativePath
	directoryPath := fileChange.DirectoryPath

	from := snapshot.FilePath(directoryPath)
	to := filepath.Join(fileChange.WatchDirectoryName)
	log.Printf("!!!! gdrive.SyncSnapshot(): sync snapshot from [%s] to [gdrive:%s]\n", from, to)
	s.store(from, to, "application/octet-stream")
}

// gdrive.store(): C:\Users\Brown\MyFiles\pixiv\71738080_p0_master1200.jpg > MyFiles\pixiv
func (s *Storage) store(file, toPath, mimeType string) {
	sleepRandom()
	log.Printf(">>> gdrive.store(): [%s] > [gdrive://%s]\n", file, toPath)
	from, err := os.Open(file)
	if err != nil {
		log.Fatalf("[ERROR] gdrive.store(): Cannot open file  [%s]: %v\n", file, err)
		return
	}
	defer from.Close()
	lastFolder := s.CreateAllFolders(toPath) // TODO: errors?

	f := &drive.File{
		Name:     filepath.Base(file),
		MimeType: mimeType, //"image/jpeg", // TODO: get the mimetype
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
