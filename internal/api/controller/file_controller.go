package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ADRIFT-backend/internal/dto"
	myerror "ADRIFT-backend/internal/pkg/error"
	"ADRIFT-backend/internal/pkg/response"
	"ADRIFT-backend/internal/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileController interface {
	ServeUpload(ctx *gin.Context)
	UploadFRSTempFile(ctx *gin.Context)
}

type fileController struct {
	storage storage.FileSystemStorage
}

func NewFileController(storage storage.FileSystemStorage) FileController {
	return &fileController{storage: storage}
}

func (c *fileController) ServeUpload(ctx *gin.Context) {
	rawPath := strings.TrimSpace(ctx.Param("path"))
	if rawPath == "" || rawPath == "/" {
		response.NewFailed("invalid path", myerror.New("path is required", http.StatusBadRequest)).Send(ctx)
		return
	}

	cleanPath := filepath.Clean(strings.TrimPrefix(rawPath, "/"))
	if cleanPath == "." || filepath.IsAbs(cleanPath) {
		response.NewFailed("invalid path", myerror.New("invalid path", http.StatusBadRequest)).Send(ctx)
		return
	}

	baseDirAbs, err := filepath.Abs("./assets")
	if err != nil {
		response.NewFailed("internal server error", myerror.New("failed to resolve assets path", http.StatusInternalServerError)).Send(ctx)
		return
	}

	targetPath := filepath.Join(baseDirAbs, cleanPath)
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		response.NewFailed("invalid path", myerror.New("failed to resolve file path", http.StatusBadRequest)).Send(ctx)
		return
	}

	relPath, err := filepath.Rel(baseDirAbs, targetPathAbs)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		response.NewFailed("invalid path", myerror.New("path traversal detected", http.StatusBadRequest)).Send(ctx)
		return
	}

	fileInfo, err := os.Stat(targetPathAbs)
	if os.IsNotExist(err) {
		response.NewFailed("file not found", myerror.New("file not found", http.StatusNotFound)).Send(ctx)
		return
	}
	if err != nil {
		response.NewFailed("internal server error", myerror.New("failed to access file", http.StatusInternalServerError)).Send(ctx)
		return
	}
	if fileInfo.IsDir() {
		response.NewFailed("file not found", myerror.New("file not found", http.StatusNotFound)).Send(ctx)
		return
	}

	ctx.File(targetPathAbs)
}

func (c *fileController) UploadFRSTempFile(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPLOAD_FRS, myerror.New("file is required", http.StatusBadRequest)).Send(ctx)
		return
	}

	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	if fileExt != ".xlsx" && fileExt != ".xls" {
		response.NewFailed(dto.MESSAGE_FAILED_UPLOAD_FRS, dto.ErrInvalidExcelFile).Send(ctx)
		return
	}

	fileName := fmt.Sprintf("frs_tmp_%s%s", uuid.New().String(), fileExt)
	allowedMimes := []string{
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-excel",
		"application/zip",
	}

	objectKey, err := c.storage.UploadFile(fileName, file, "tmp/frs", allowedMimes...)
	if err != nil {
		response.NewFailed(dto.MESSAGE_FAILED_UPLOAD_FRS, err, nil).Send(ctx)
		return
	}

	response.NewSuccess(dto.MESSAGE_SUCCESS_UPLOAD_FRS, dto.FRSTempUploadResponse{
		ObjectKey: objectKey,
		FileName:  fileName,
	}).Send(ctx)
}
