package gdrive

import (
	"fmt"
	"log"
	"os"

	drive "google.golang.org/api/drive/v3"
)

// FindFile ...
// For testing: https://developers.google.com/drive/api/v3/reference/files/list
func (s *Storage) FindFile(name, folderID string) (*drive.File, error) {
	q := fmt.Sprintf("mimeType != 'application/vnd.google-apps.folder' and trashed = false and name = '%s' and '%s' in parents", name, folderID)
	files, err := s.service.Files.List().Q(q).Do()

	if err != nil {
		return nil, err
	}

	if len(files.Files) == 0 {
		return nil, nil
	}

	if len(files.Files) == 1 {
		return files.Files[0], nil
	}
	
	for _, file := range files.Files {
		log.Printf(">>>\t\tName: %s, ID: %s\n", file.Name, file.Id)
	}

	return nil, fmt.Errorf("gdrive.FindFile(): Too many files (%d) found", len(files.Files))
}

// CreateOrUpdateFile ...
func (s *Storage) CreateOrUpdateFile(fromFile *os.File, fileName, mimeType, folderID string) (*drive.File, error) {
	mu.Lock()
	defer mu.Unlock()

	// DefaultUploadChunkSize = 8 * 1024 * 1024
	// chunkSize := googleapi.ChunkSize(5 * 1024 * 1024)
	// contentType := googleapi.ContentType(mimeType)

	file, err := s.FindFile(fileName, folderID)
	if err == nil && file == nil { // file was not found, no error no file
		f := &drive.File{
			Name:     fileName,
			MimeType: mimeType,
			Parents:  []string{folderID},
		}
		file, err = s.service.Files.Create(f).Media(fromFile).Do()
		return file, err
	}

	if err == nil && file != nil {
		file, err = s.service.Files.Update(file.Id, nil).Media(fromFile).Do()
		return file, err
	}

	return nil, fmt.Errorf("cannot create or update file on GDrive: %s", fileName)
}
