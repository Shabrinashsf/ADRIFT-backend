package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"

	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/pkg/validate"
)

type (
	FileSystemStorage interface {
		UploadFile(filename string, file *multipart.FileHeader, folderName string, mv ...string) (string, error)
		DeleteFile(objectKey string) error
		GetPublicLink(objectKey string) string
		Begin() FileSystemStorage
		Commit()
		Rollback()
	}

	actionOption struct {
		actionType string
		key        string
	}

	filesystemStorage struct {
		uploadPath string
		actions    []actionOption
		isRollback bool
	}
)

// NewFileSystemStorage creates a new filesystem storage service
func NewFileSystemStorage() FileSystemStorage {
	uploadPath := "./assets"
	return &filesystemStorage{
		uploadPath: uploadPath,
		actions:    nil,
		isRollback: false,
	}
}

// UploadFile saves a file to the filesystem and returns the object key
func (fs *filesystemStorage) UploadFile(filename string, file *multipart.FileHeader, folderName string, mv ...string) (string, error) {
	uploadedFile, err := file.Open()
	if err != nil {
		return "", err
	}

	mimetype, err := validate.GetMimeType(uploadedFile)
	uploadedFile.Close()
	if err != nil {
		return "", err
	}

	if len(mv) > 0 {
		flag := false
		for _, m := range mv {
			if mimetype == m {
				flag = true
				break
			}
		}
		if !flag {
			return "", dto.ErrInvalidTypeFile
		}
	}

	safeFilename := fmt.Sprintf("%s", filename)

	folderPath := filepath.Join(fs.uploadPath, folderName)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	fullPath := filepath.Join(folderPath, safeFilename)

	uploadedFile, err = file.Open()
	if err != nil {
		return "", err
	}
	defer uploadedFile.Close()

	destFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, uploadedFile); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	objectKey := fmt.Sprintf("%s/%s", folderName, safeFilename)

	if fs.isRollback {
		fs.actions = append(fs.actions, actionOption{actionType: "upload", key: objectKey})
	}

	return objectKey, nil
}

// DeleteFile removes a file from the filesystem
func (fs *filesystemStorage) DeleteFile(objectKey string) error {
	fullPath := filepath.Join(fs.uploadPath, objectKey)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetPublicLink generates an HTTP-accessible URL for the file
func (fs *filesystemStorage) GetPublicLink(objectKey string) string {
	return fmt.Sprintf("/api/assets/%s", objectKey)
}

// Begin starts a transaction for potential rollback
func (fs *filesystemStorage) Begin() FileSystemStorage {
	fs.actions = []actionOption{}
	fs.isRollback = true
	return fs
}

// Commit marks the transaction as complete
func (fs *filesystemStorage) Commit() {
	fs.actions = nil
	fs.isRollback = false
}

// Rollback deletes files in reverse order (last in, first out)
func (fs *filesystemStorage) Rollback() {
	var wg sync.WaitGroup
	errCh := make(chan error, len(fs.actions))

	for i := len(fs.actions) - 1; i >= 0; i-- {
		action := fs.actions[i]
		switch action.actionType {
		case "upload":
			wg.Add(1)
			go func(key string) {
				defer wg.Done()
				if err := fs.DeleteFile(key); err != nil {
					errCh <- fmt.Errorf("failed to delete file %s: %v", key, err)
				}
			}(action.key)
		}
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		panic("failed rollback: " + errors[0].Error())
	}

	fs.Commit()
}
