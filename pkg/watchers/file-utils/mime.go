package fileutils

import (
	"net/http"
	"os"
)

// ContentType returns mime type of the file as a string
// source: https://golangcode.com/get-the-content-type-of-file/
func ContentType(filePath string) (string, error) {
	out, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

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
