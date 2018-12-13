package gdrive

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	drive "google.golang.org/api/drive/v3"
)

var mu sync.Mutex

// CreateFolder Creates a new folder in gdrive
func (s *Storage) CreateFolder(name string) *drive.File {
	log.Printf("gdrive.CreateFolder(): name=%s\n", name)

	folder, err := s.FindFolder(name, &FindFileOptions{})
	if err != nil {
		log.Panicf("[ERROR] gdrive.CreateFolder(): Unable to create folder [%s]: %v\n", name, err)
	}
	if err == nil && folder != nil {
		return folder
	}

	createFolder, err := s.service.Files.Create(&drive.File{
		Name: name, MimeType: "application/vnd.google-apps.folder",
		FolderColorRgb: "7FB069", // ASPARAGUS
	}).Do()
	if err != nil {
		log.Panicf("[ERROR] gdrive.CreateFolder(): Unable to create folder [%s]: %v\n", name, err)
	}
	return createFolder
}

// FindOrCreateSubFolder ...
func (s *Storage) FindOrCreateSubFolder(parentFolderID, name string) *drive.File {
	log.Printf("gdrive.CreateSubFolder(): parentID=%s, name=%s\n", parentFolderID, name)

	folder, err := s.FindFolder(name, &FindFileOptions{ParentFolderID: parentFolderID})
	if err != nil {
		log.Panicf("[ERROR] gdrive.FindOrCreateSubFolder(): Unable to create folder [%s]: %v\n", name, err)
	}
	if err == nil && folder != nil {
		return folder
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
	mu.Lock()
	defer mu.Unlock()
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

type FindFileOptions struct {
	ParentFolderID string
}

// FindFolder returns a folder by name if a single folder is found or an error. If no folder is found, return nil
func (s *Storage) FindFolder(name string, params *FindFileOptions) (*drive.File, error) {
	q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s'", name)
	if params.ParentFolderID != "" {
		q = fmt.Sprintf("%s and '%s' in parents", q, params.ParentFolderID)
	}

	// q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents", name, parentFolderID)
	folders, err := s.service.Files.List().Q(q).Do()
	if err != nil {
		return nil, err
	}

	if len(folders.Files) == 1 {
		log.Printf(">>> gdrive.FindFolder(): folder [%s] found\n", name)
		return folders.Files[0], nil
	}
	if len(folders.Files) > 1 {
		for _, folder := range folders.Files {
			log.Printf(">>> 	Name: %s, ID: %s\n", folder.Name, folder.Id)
		}

		return nil, fmt.Errorf("gdrive.CreateFolder(): Too many folders found")
	}
	return nil, nil
}
