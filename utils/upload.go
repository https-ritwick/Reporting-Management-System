package utils

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func SaveUploadedFile(r *http.Request, fieldName string, appNumber string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", nil // Field might be optional
	}
	defer file.Close()

	// Ensure uploads directory exists
	dir := "static/uploads"
	os.MkdirAll(dir, os.ModePerm)

	filename := appNumber + "_" + fieldName + filepath.Ext(header.Filename)
	filePath := filepath.Join(dir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return "/" + filePath, nil // Return relative path for DB
}
