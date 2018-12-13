package gdrive

import (
	"fmt"
	"log"
	"os"
	"strings"

	drive "google.golang.org/api/drive/v3"
)

// CreateFolder Creates a new folder in gdrive
func (s *Storage) CreateFolder(name string) *drive.File {
	log.Printf("gdrive.CreateFolder(): name=%s\n", name)

	q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s'", name)
	folders, err := s.service.Files.List().Q(q).Do()
	if err != nil {
		log.Panicf("[ERROR] gdrive.CreateFolder(): Unable to do the query [%s]: %v\n", q, err)
	}

	if len(folders.Files) == 1 {
		return folders.Files[0]
	}
	if len(folders.Files) > 1 {
		log.Panicf("[ERROR] gdrive.CreateFolder(): Too many folders found:\n")
		for _, folder := range folders.Files {
			log.Printf("	Name: %s, ID: %s\n", folder.Name, folder.Id)
		}
		return folders.Files[0]
	}

	createFolder, err := s.service.Files.Create(&drive.File{Name: name, MimeType: "application/vnd.google-apps.folder"}).Do()
	if err != nil {
		log.Panicf("[ERROR] gdrive.CreateFolder(): Unable to create folder [%s]: %v\n", name, err)
	}
	return createFolder
}

// FindOrCreateSubFolder ...
func (s *Storage) FindOrCreateSubFolder(parentFolderID, name string) *drive.File {
	log.Printf("gdrive.CreateSubFolder(): parentID=%s, name=%s\n", parentFolderID, name)

	q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents", name, parentFolderID)
	folders, err := s.service.Files.List().Q(q).Do()
	if err != nil {
		log.Panicf("[ERROR] gdrive.CreateFolder(): Unable to do the query [%s]: %v\n", q, err)
	}

	if len(folders.Files) == 1 {
		log.Printf(">>> gdrive.FindFolder(): folder [%s] found\n", name)
		return folders.Files[0]
	}
	if len(folders.Files) > 1 {
		for _, folder := range folders.Files {
			log.Printf(">>> 	Name: %s, ID: %s\n", folder.Name, folder.Id)
		}
		log.Panicf("[ERROR] gdrive.CreateFolder(): Too many folders found!\n")
		return folders.Files[0]
	}

	createFolder, err := s.service.Files.Create(&drive.File{
		Name: name, MimeType: "application/vnd.google-apps.folder",
		Parents: []string{parentFolderID},
	}).Do()
	if err != nil {
		log.Panicf("[ERROR] gdrive.CreateSubFolder(): Unable to create folder [%s] as subfolder of [%s]: %v\n", name, parentFolderID, err)
	}
	log.Printf(">>> New folder [%s] is created\n", name)
	return createFolder
}

// CreateAllFolders creates all folders from the path like /foo/bar/buz
// and returns last folder in the path
func (s *Storage) CreateAllFolders(path string) *drive.File {
	log.Printf("gdirve.CreateAllFolders(): %s\n", path)
	paths := strings.Split(path, string(os.PathSeparator))
	parentID := s.root.Id
	var f *drive.File
	for _, name := range paths {
		f = s.FindOrCreateSubFolder(parentID, name)
		parentID = f.Id
	}
	return f
}
