package dto

import (
	"mime/multipart"
	"net/http"

	myerror "ADRIFT-backend/internal/pkg/error"
)

const (
	// Failed
	MESSAGE_FAILED_UPLOAD_FRS  = "gagal mengupload file frs"
	MESSAGE_FAILED_PROCESS_FRS = "gagal memproses file frs"
	MESSAGE_FAILED_DELETE_FRS  = "gagal menghapus file frs"
	MESSAGE_FAILED_SUBMIT_FRS  = "gagal submit data frs"

	// Success
	MESSAGE_SUCCESS_UPLOAD_FRS  = "berhasil mengupload file frs"
	MESSAGE_SUCCESS_PROCESS_FRS = "berhasil memproses file frs"
	MESSAGE_SUCCESS_DELETE_FRS  = "berhasil menghapus file frs"
	MESSAGE_SUCCESS_SUBMIT_FRS  = "berhasil submit data frs"
)

var (
	ErrInvalidAcademicYear  = myerror.New("tahun ajaran harus format YYYY/YYYY", http.StatusBadRequest)
	ErrInvalidTerm          = myerror.New("term harus ganjil atau genap", http.StatusBadRequest)
	ErrInvalidExcelFile     = myerror.New("file harus berupa excel (.xlsx atau .xls)", http.StatusBadRequest)
	ErrScheduleAlreadyExist = myerror.New("data jadwal untuk tahun ajaran dan term tersebut sudah ada", http.StatusConflict)
	ErrScheduleFileNotFound = myerror.New("file schedule tidak ditemukan", http.StatusNotFound)
	ErrTempFileNotFound     = myerror.New("file temporary tidak ditemukan", http.StatusNotFound)
)

type (
	FRSTempUploadRequest struct {
		File *multipart.FileHeader `form:"file" binding:"required"`
	}

	FRSTempUploadResponse struct {
		ObjectKey string `json:"object_key"`
		FileName  string `json:"file_name"`
	}

	FRSUploadRequest struct {
		ObjectKey    string `json:"object_key" form:"object_key" binding:"required"`
		AcademicYear string `json:"academic_year" form:"academic_year" binding:"required"`
		Term         string `json:"term" form:"term" binding:"required"`
	}

	FRSUploadResponse struct {
		FileURL      string           `json:"file_url"`
		ObjectKey    string           `json:"object_key"`
		FileName     string           `json:"file_name"`
		AcademicYear string           `json:"academic_year"`
		Term         string           `json:"term"`
		NullRecords  []map[string]any `json:"null_records"`
	}

	FRSUploadDeleteRequest struct {
		ObjectKey string `json:"object_key" form:"object_key" binding:"required"`
	}

	FRSSubmitRequest struct {
		AcademicYear string `json:"academic_year" form:"academic_year" binding:"required"`
		Term         string `json:"term" form:"term" binding:"required"`
	}

	FRSSubmitResponse struct {
		Inserted int `json:"inserted"`
	}
)
