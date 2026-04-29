package dto

import (
	"mime/multipart"
	"net/http"
	"time"

	"ADRIFT-backend/internal/entity"
	myerror "ADRIFT-backend/internal/pkg/error"
)

const (
	// Failed
	MESSAGE_FAILED_UPLOAD_FRS      = "gagal mengupload file frs"
	MESSAGE_FAILED_PROCESS_FRS     = "gagal memproses file frs"
	MESSAGE_FAILED_DELETE_FRS      = "gagal menghapus file frs"
	MESSAGE_FAILED_SUBMIT_FRS      = "gagal submit data frs"
	MESSAGE_FAILED_LIST_SCHEDULES   = "gagal mendapatkan daftar jadwal"
	MESSAGE_FAILED_CREATE_PLAN      = "gagal membuat rencana frs"
	MESSAGE_FAILED_LIST_PLANS       = "gagal mendapatkan daftar rencana frs"
	MESSAGE_FAILED_GET_PLAN_DETAIL  = "gagal mendapatkan detail rencana frs"
	MESSAGE_FAILED_FIND_ALTERNATIVE = "gagal mencari jadwal alternatif"
	MESSAGE_FAILED_DELETE_PLAN        = "gagal menghapus rencana frs"

	// Success
	MESSAGE_SUCCESS_UPLOAD_FRS      = "berhasil mengupload file frs"
	MESSAGE_SUCCESS_PROCESS_FRS     = "berhasil memproses file frs"
	MESSAGE_SUCCESS_DELETE_FRS      = "berhasil menghapus file frs"
	MESSAGE_SUCCESS_SUBMIT_FRS      = "berhasil submit data frs"
	MESSAGE_SUCCESS_LIST_SCHEDULES  = "berhasil mendapatkan daftar jadwal"
	MESSAGE_SUCCESS_CREATE_PLAN     = "berhasil membuat rencana frs"
	MESSAGE_SUCCESS_LIST_PLANS      = "berhasil mendapatkan daftar rencana frs"
	MESSAGE_SUCCESS_GET_PLAN_DETAIL = "berhasil mendapatkan detail rencana frs"
	MESSAGE_SUCCESS_FIND_ALTERNATIVE = "berhasil mendapatkan jadwal alternatif"
	MESSAGE_SUCCESS_DELETE_PLAN       = "berhasil menghapus rencana frs"
)

var (
	ErrInvalidAcademicYear       = myerror.New("tahun ajaran harus format YYYY/YYYY", http.StatusBadRequest)
	ErrInvalidTerm               = myerror.New("term harus ganjil atau genap", http.StatusBadRequest)
	ErrInvalidExcelFile          = myerror.New("file harus berupa excel (.xlsx atau .xls)", http.StatusBadRequest)
	ErrScheduleAlreadyExist      = myerror.New("data jadwal untuk tahun ajaran dan term tersebut sudah ada", http.StatusConflict)
	ErrScheduleFileNotFound      = myerror.New("file schedule tidak ditemukan", http.StatusNotFound)
	ErrTempFileNotFound          = myerror.New("file temporary tidak ditemukan", http.StatusNotFound)
	ErrPlanNotFound              = myerror.New("rencana frs tidak ditemukan", http.StatusNotFound)
	ErrPlanNotOwnedByUser        = myerror.New("rencana frs bukan milik user ini", http.StatusForbidden)
	ErrScheduleIDNotFound        = myerror.New("salah satu schedule_id tidak ditemukan", http.StatusNotFound)
	ErrNoAlternativeFound        = myerror.New("tidak ditemukan jadwal alternatif", http.StatusNotFound)
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
		FileURL             string           `json:"file_url"`
		ObjectKey           string           `json:"object_key"`
		FileName            string           `json:"file_name"`
		AcademicYear        string           `json:"academic_year"`
		Term                string           `json:"term"`
		NullRecords         []map[string]any `json:"null_records"`
		MissingLectureCodes []string         `json:"missing_lecture_codes,omitempty"`
	}

	FRSUploadDeleteRequest struct {
		ObjectKey string `json:"object_key" form:"object_key" binding:"required"`
	}

	FRSSubmitRequest struct {
		ObjectKey    string `json:"object_key" form:"object_key" binding:"required"`
		AcademicYear string `json:"academic_year" form:"academic_year" binding:"required"`
		Term         string `json:"term" form:"term" binding:"required"`
	}

	FRSSubmitResponse struct {
		Inserted int `json:"inserted"`
	}

	ScheduleResponse struct {
		ID          string           `json:"id"`
		CourseName  string           `json:"course_name"`
		SKS         int              `json:"sks"`
		Class       string           `json:"class"`
		Day         entity.Day       `json:"day"`
		StartTime   string           `json:"start_time"`
		EndTime     string           `json:"end_time"`
		LectureID   string           `json:"lecture_id"`
		LectureName string           `json:"lecture_name"`
		Room        string           `json:"room"`
		Capacity    int              `json:"capacity"`
		Semester    int              `json:"semester"`
		Prodi       entity.ProdiType `json:"prodi"`
	}

	// Create FRS Plan
	CreateFRSPlanRequest struct {
		PlanName      string   `json:"plan_name" binding:"required"`
		AcademicYear  string   `json:"academic_year" binding:"required"`
		Term          string   `json:"term" binding:"required"`
		TotalCredit   int      `json:"total_credit" binding:"required"`
		ScheduleIDs   []string `json:"schedule_ids" binding:"required,min=1"`
	}

	// Alternative schedules
	AlternativeScheduleRequest struct {
		PlanName      string   `json:"plan_name" binding:"required"`
		AcademicYear  string   `json:"academic_year" binding:"required"`
		Term          string   `json:"term" binding:"required"`
		TotalCredit   int      `json:"total_credit" binding:"required"`
		ScheduleIDs   []string `json:"schedule_ids" binding:"required,min=1"`
	}

	AlternativeScheduleResponse struct {
		Alternatives []AlternativeGroup `json:"alternatives"`
	}

	AlternativeGroup struct {
		PriorityNote string                   `json:"priority_note"`
		Schedules    []AlternativeScheduleItem `json:"schedules"`
	}

	AlternativeScheduleItem struct {
		ScheduleID  string `json:"schedule_id"`
		CourseName  string `json:"course_name"`
		Class       string `json:"class"`
		LectureName string `json:"lecture_name"`
		Day         string `json:"day"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
		SKS         int    `json:"sks"`
	}

	// List plans
	FRSPlanListItem struct {
		ID           string    `json:"id"`
		PlanName     string    `json:"plan_name"`
		AcademicYear string    `json:"academic_year"`
		Term         string    `json:"term"`
		TotalCredit  int       `json:"total_credit"`
		CourseCount  int       `json:"course_count"`
		CreatedAt    time.Time `json:"created_at"`
	}

	// Plan detail
	FRSPlanDetailResponse struct {
		ID           string              `json:"id"`
		PlanName     string              `json:"plan_name"`
		AcademicYear string              `json:"academic_year"`
		Term         string              `json:"term"`
		TotalCredit  int                 `json:"total_credit"`
		Items        []FRSPlanItemDetail `json:"items"`
	}

	FRSPlanItemDetail struct {
		ID          string `json:"id"`
		ScheduleID  string `json:"schedule_id"`
		CourseName  string `json:"course_name"`
		Class       string `json:"class"`
		Day         string `json:"day"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		LectureName string `json:"lecture_name"`
		Room        string `json:"room"`
		Credit      int    `json:"credit"`
	}
)
