package gdrive

import (
	"fmt"
	"log"
	"os"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// FindFile ...
func (s *Storage) FindFile(name, folderID, mimeType string) (*drive.File, error) {
	fmt.Printf(">>> FindFile(): %s [%s]\n", name, mimeType)

	q := fmt.Sprintf("mimeType = '%s' and name = '%s' and '%s' in parents", mimeType, name, folderID)
	files, err := s.service.Files.List().Q(q).Do()

	if err != nil {
		fmt.Printf(">>> gdrive.FindFile(): file [%s] not found [%v]\n", name, err)
		fmt.Printf("q=%s\n", q)
		return nil, err
	}

	fmt.Printf("!!!!!! gdrive.FindFile(): files total [%d]\n", len(files.Files))

	if len(files.Files) == 0 {
		fmt.Printf(">>> gdrive.FindFile(): file [%s] not found in [%s]\n", name, folderID)
		fmt.Printf("q=%s\n", q)
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

	f := &drive.File{
		Name:     fileName,
		MimeType: mimeType,
		Parents:  []string{folderID},
	}

	// DefaultUploadChunkSize = 8 * 1024 * 1024
	chunkSize := googleapi.ChunkSize(5 * 1024 * 1024)
	contentType := googleapi.ContentType(mimeType)

	// TODO: wrap this with createOrUpdateFile
	file, err := s.FindFile(fileName, folderID, mimeType)
	if err == nil && file == nil { // file was not found, no error no file
		file, err = s.service.Files.Create(f).Media(fromFile, chunkSize, contentType).Do()
		return file, err
	}

	if err == nil && file != nil {
		file, err = s.service.Files.Update("", f).Media(fromFile, chunkSize, contentType).Do()
		return file, err
	}

	return nil, fmt.Errorf("cannot create or update file on GDrive: %s", fileName)
}
