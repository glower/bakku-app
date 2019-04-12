package file

import (
	"crypto/sha512"
	"fmt"
	"io"
	"net/http"
	"os"
)

type FileInformation struct {
	absoluteFilePath string
}
type FileInformationImplementer interface {
	ContentType() (string, error)
	Checksum() (string, error)
}

// ExtendedFileInfo is combined receiver for os.FileInfo functions and ContentType()
type ExtendedFileInfo struct {
	FileInformation
	os.FileInfo
}

type ExtendedFileInfoImplementer interface {
	os.FileInfo
	FileInformationImplementer
}

func ExtendedFileInformation(absoluteFilePath string, fileInfo os.FileInfo) ExtendedFileInfoImplementer {
	return &ExtendedFileInfo{
		FileInfo: fileInfo,
		FileInformation: FileInformation{
			absoluteFilePath: absoluteFilePath,
		},
	}
}

func GetFileInformation(absoluteFilePath string) (ExtendedFileInfoImplementer, error) {

	fileInfo, err := os.Stat(absoluteFilePath)
	if err != nil {
		return nil, fmt.Errorf("can't stat file [%s]: %v", absoluteFilePath, err)
	}

	return &ExtendedFileInfo{
		FileInfo: fileInfo,
		FileInformation: FileInformation{
			absoluteFilePath: absoluteFilePath,
		},
	}, nil
}

// ContentType returns mime type of the file as a string
// source: https://golangcode.com/get-the-content-type-of-file/
func (i *FileInformation) ContentType() (string, error) {
	out, err := os.Open(i.absoluteFilePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

// Checksum returns a string representation of SHA-512/256 checksum
func (i *FileInformation) Checksum() (string, error) {
	f, err := os.Open(i.absoluteFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha512.New512_256()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
