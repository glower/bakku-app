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
	fmt.Printf(">>> FindFile(): %s\n", name)

	q := fmt.Sprintf("mimeType != 'application/vnd.google-apps.folder' and trashed = false and name = '%s' and '%s' in parents", name, folderID)
	files, err := s.service.Files.List().Q(q).Do()

	if err != nil {
		fmt.Printf(">>> gdrive.FindFile(): file [%s] not found [%v]\n", name, err)
		fmt.Printf("q=%s\n", q)
		return nil, err
	}

	if len(files.Files) == 0 {
		fmt.Printf(">>> gdrive.FindFile(): file [%s] not found in [%s]\n", name, folderID)
		return nil, nil
	}

	if len(files.Files) == 1 {
		fmt.Printf(">>> gdrive.FindFile(): found 1 file [%s]\n", files.Files[0].Name)
		return files.Files[0], nil
	}
	for _, file := range files.Files {
		log.Printf(">>>\t\tName: %s, ID: %s\n", file.Name, file.Id)
	}
	return nil, fmt.Errorf("gdrive.CreateFolder(): Too many folders found")
}

// CreateOrUpdateFile ...
func (s *Storage) CreateOrUpdateFile(fromFile *os.File, fileName, mimeType, folderID string) (*drive.File, error) {
	fmt.Printf(">>> CreateOrUpdateFile(): %s\n", fileName)
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
		log.Printf("gdrive.CreateOrUpdateFile(): try to update file [%s] with id [%s]\n", file.Name, file.Id)
		file, err = s.service.Files.Update(file.Id, nil).Media(fromFile).Do()
		return file, err
	}

	return nil, fmt.Errorf("cannot create or update file on GDrive: %s", fileName)
}
