package utils

import (
	"encoding/base64"
	"io"
	"os"
)

// FileToBase64 reads a file from the given path and returns it as base64 encoded bytes
func FileToBase64(filePath string) ([]byte, error) {
	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Encode to base64
	base64String := base64.StdEncoding.EncodeToString(fileContent)
	
	// Convert the base64 string to bytes
	return []byte(base64String), nil
}
