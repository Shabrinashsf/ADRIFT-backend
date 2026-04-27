package validate

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"errors"

	"ADRIFT-backend/internal/pkg/logger"
)

var (
	AvailableMimeTypes = map[string][]string{
		// images
		"image/jpeg": {".jpg", ".jpeg"},
		"image/jpg":  {".jpg", ".jpeg"},
		"image/png":  {".png"},

		// documents
		"application/pdf":          {".pdf"},
		"application/octet-stream": {".pdf"},
	}

	AllowImage    = []string{"image/jpeg", "image/jpg", "image/png"}
	AllowImagePdf = []string{"image/jpeg", "image/jpg", "image/png", "application/pdf"}
)

var (
	ErrInvalidSize = "%s: invalid file size, must be less than %d mb"
	ErrInvalidType = "%s: invalid file type, must be %s"
	ErrInvalidExt  = "%s: invalid file extension mismatch with file type"
)

func ValidateFile(f *multipart.FileHeader, size int64, mimes ...string) ([]byte, string, error) {
	// check file size
	if f == nil || f.Size == 0 {
		return nil, "", errors.New("file is empty")
	}

	if f.Size > size {
		sizeInMb := size / 1024 / 1024
		return nil, "", errors.New(fmt.Sprintf(ErrInvalidSize, f.Filename, sizeInMb))
	}

	file, err := f.Open()
	if err != nil {
		return nil, "", errors.New("failed to open file")
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("failed to close file: %v", err)
		}
	}()

	// check file type
	ext := strings.ToLower(filepath.Ext(f.Filename))
	if len(mimes) > 0 {
		if ext == "" {
			return nil, "", errors.New(fmt.Sprintf(ErrInvalidExt, f.Filename))
		}

		mime, err := GetMimeType(file)
		if err != nil {
			return nil, "", err
		}

		flag := false
		for _, m := range mimes {
			if m == strings.ToLower(mime) {
				expectedExts := AvailableMimeTypes[m]
				if !slices.Contains(expectedExts, ext) {
					return nil, "", errors.New(fmt.Sprintf(ErrInvalidExt, f.Filename))
				}
				flag = true
				break
			}
		}
		if !flag {
			return nil, "", errors.New(fmt.Sprintf(ErrInvalidType, f.Filename, strings.Join(mimes, ", ")))
		}
	}

	// read actual file
	fileBytes := make([]byte, f.Size)
	if _, err := file.Read(fileBytes); err != nil {
		return nil, "", err
	}

	return fileBytes, ext, nil
}

func GetMimeType(f multipart.File) (string, error) {
	buffer := make([]byte, 512)
	if _, err := f.Read(buffer); err != nil && err != io.EOF {
		return "", err
	}

	mime := http.DetectContentType(buffer)

	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}

	return mime, nil
}

func TextToSize(size string) int64 {
	sizeType := size[len(size)-2:]
	sizeValue, _ := strconv.Atoi(size[:len(size)-2])

	switch sizeType {
	case "kb":
		return int64(sizeValue * 1024)
	case "mb":
		return int64(sizeValue * 1024 * 1024)
	case "gb":
		return int64(sizeValue * 1024 * 1024 * 1024)
	default:
		return int64(sizeValue)
	}
}
