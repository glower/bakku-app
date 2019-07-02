package gdrive

import (
	"fmt"
	"os"
	"strings"
	"sync"

	drive "google.golang.org/api/drive/v3"
)

// FindFileOptions holds options for search
type FindFileOptions struct {
	ParentFolderID string
}

var mu sync.Mutex

// CreateFolder Creates a new folder in gdrive
func (s *Storage) CreateFolder(name string) (*drive.File, error) {
	// log.Printf("gdrive.CreateFolder(): name=%s\n", name)
	folder, err := s.FindFolder(name, &FindFileOptions{})
	if err != nil {
		return nil, fmt.Errorf("gdrive.CreateFolder(): Unable to create folder [%s]: %v", name, err)
	}
	if err == nil && folder != nil {
		return folder, nil
	}

	if err == nil && folder == nil {
		createFolder, err := s.service.Files.Create(&drive.File{
			Name: name, MimeType: "application/vnd.google-apps.folder",
			FolderColorRgb: "7FB069", // ASPARAGUS
		}).Do()

		if err != nil {
			return nil, fmt.Errorf("gdrive.CreateFolder(): Unable to create folder [%s]: %v", name, err)
		}
		return createFolder, nil
	}

	return nil, fmt.Errorf("gdrive.CreateFolder(): Unable to create folder [%s]", name)
}

// FindOrCreateSubFolder ...
func (s *Storage) FindOrCreateSubFolder(parentFolderID, name string) (*drive.File, error) {
	// log.Printf("gdrive.CreateSubFolder(): parentID=%s, name=%s\n", parentFolderID, name)
	folder, err := s.FindFolder(name, &FindFileOptions{ParentFolderID: parentFolderID})
	if err != nil {
		return nil, fmt.Errorf("gdrive.FindOrCreateSubFolder(): Unable to create folder [%s]: %v", name, err)
	}
	if err == nil && folder != nil {
		// fmt.Printf("SubFolder(): folder found: [%s] [id:%s]\n", folder.Name, folder.Id)
		return folder, nil
	}

	createFolder, err := s.service.Files.Create(&drive.File{
		Name: name, MimeType: "application/vnd.google-apps.folder",
		Parents: []string{parentFolderID},
	}).Do()

	if err != nil {
		return nil, fmt.Errorf("gdrive.CreateSubFolder(): Unable to create folder [%s] as subfolder of [%s]: %v", name, parentFolderID, err)
	}
	return createFolder, nil
}

// GetOrCreateAllFolders creates all folders from the path like /foo/bar/buz
// and returns last folder in the path
func (s *Storage) GetOrCreateAllFolders(path string) (*drive.File, error) {
	mu.Lock()
	defer mu.Unlock()
	// log.Printf("gdirve.CreateAllFolders(): %s\n", path)
	paths := strings.Split(path, string(os.PathSeparator))
	parentID := s.root.Id
	var f *drive.File
	var err error
	for _, name := range paths {
		f, err = s.FindOrCreateSubFolder(parentID, name)
		if err != nil {
			return nil, err
		}
		parentID = f.Id
	}
	// fmt.Printf("GetOrCreateAllFolders(): path=%v, file=%s\n", paths, f.Name)
	return f, err
}

// FindFolder returns a folder by name if a single folder is found or an error. If no folder is found, return nil
func (s *Storage) FindFolder(name string, params *FindFileOptions) (*drive.File, error) {
	q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and trashed = false and name = '%s'", name)
	if params.ParentFolderID != "" {
		q = fmt.Sprintf("%s and '%s' in parents", q, params.ParentFolderID)
	}

	folders, err := s.service.Files.List().Q(q).Do()
	if err != nil {
		return nil, err
	}

	if len(folders.Files) > 1 {
		return nil, fmt.Errorf("gdrive.CreateFolder(): Too many folders [%s] (%d) found", name, len(folders.Files))
	}

	if len(folders.Files) == 1 {
		return folders.Files[0], nil
	}

	return nil, nil
}
