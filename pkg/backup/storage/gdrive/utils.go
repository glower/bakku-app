package gdrive

import (
	"log"
	"os"
	"strings"

	drive "google.golang.org/api/drive/v3"
)

// CreateFolder Creates a new folder in gdrive
func (s *Storage) CreateFolder(name string) *drive.File {
	log.Printf("gdrive.CreateFolder(): name=%s\n", name)

	createFolder, err := s.service.Files.Create(&drive.File{Name: name, MimeType: "application/vnd.google-apps.folder"}).Do()
	if err != nil {
		log.Printf("[ERROR] gdrive.CreateFolder(): Unable to create folder [%s]: %v", name, err)
	}
	return createFolder
}

// CreateSubFolder ...
func (s *Storage) CreateSubFolder(parentFolderID, name string) *drive.File {
	log.Printf("gdrive.CreateSubFolder(): parentID=%s, name=%s\n", parentFolderID, name)
	createFolder, err := s.service.Files.Create(&drive.File{
		Name: name, MimeType: "application/vnd.google-apps.folder",
		Parents: []string{parentFolderID},
	}).Do()
	if err != nil {
		log.Printf("[ERROR] gdrive.CreateSubFolder(): Unable to create folder [%s] as subfolder of [%s]: %v", name, parentFolderID, err)
	}
	return createFolder
}

// CreateAllFolders creates all folders from the path like /foo/bar/buz
// and returns last folder in the path
func (s *Storage) CreateAllFolders(path string) *drive.File {
	log.Printf("gdirve.CreateAllFolders(): %s\n", path)
	paths := strings.Split(path, string(os.PathSeparator))
	log.Printf("gdirve.CreateAllFolders(): paths=%v\n", paths)
	parentID := s.root.Id
	log.Printf("gdirve.CreateAllFolders(): rootID=%v\n", parentID)
	var f *drive.File
	for _, name := range paths {
		f = s.CreateSubFolder(parentID, name)
		parentID = f.Id
	}
	return f
}
