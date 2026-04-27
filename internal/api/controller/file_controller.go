package controller

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	myerror "ADRIFT-backend/internal/pkg/error"
	"ADRIFT-backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type FileController interface {
	ServeUpload(ctx *gin.Context)
}

type fileController struct{}

func NewFileController() FileController {
	return &fileController{}
}

func (c *fileController) ServeUpload(ctx *gin.Context) {
	rawPath := strings.TrimSpace(ctx.Param("path"))
	if rawPath == "" || rawPath == "/" {
		response.NewFailed("invalid path", myerror.New("path is required", http.StatusBadRequest)).Send(ctx)
		return
	}

	// Gin wildcard path usually starts with '/'. Keep it relative to the assets directory.
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
