package dto

import (
	"net/http"

	myerror "ADRIFT-backend/internal/pkg/error"
)

const (
	MESSAGE_SUCCESS_LIST_COURSE_GROUPS    = "berhasil mendapatkan daftar course per semester"
	MESSAGE_SUCCESS_LIST_COURSES_SEMESTER = "berhasil mendapatkan daftar course"
	MESSAGE_SUCCESS_CREATE_COURSE         = "berhasil menambahkan course"
	MESSAGE_SUCCESS_UPDATE_COURSE         = "berhasil mengupdate course"
	MESSAGE_SUCCESS_DELETE_COURSE         = "berhasil menghapus course"
	MESSAGE_SUCCESS_LIST_SCHEDULE_GROUPS  = "berhasil mendapatkan daftar schedule"
	MESSAGE_SUCCESS_LIST_SCHEDULES_FILTER = "berhasil mendapatkan daftar schedule"
	MESSAGE_SUCCESS_CREATE_SCHEDULE       = "berhasil menambahkan schedule"
	MESSAGE_SUCCESS_UPDATE_SCHEDULE       = "berhasil mengupdate schedule"
	MESSAGE_SUCCESS_DELETE_SCHEDULE       = "berhasil menghapus schedule"

	MESSAGE_FAILED_LIST_COURSE_GROUPS    = "gagal mendapatkan daftar course per semester"
	MESSAGE_FAILED_LIST_COURSES_SEMESTER = "gagal mendapatkan daftar course"
	MESSAGE_FAILED_CREATE_COURSE         = "gagal menambahkan course"
	MESSAGE_FAILED_UPDATE_COURSE         = "gagal mengupdate course"
	MESSAGE_FAILED_DELETE_COURSE         = "gagal menghapus course"
	MESSAGE_FAILED_LIST_SCHEDULE_GROUPS  = "gagal mendapatkan daftar schedule"
	MESSAGE_FAILED_LIST_SCHEDULES_FILTER = "gagal mendapatkan daftar schedule"
	MESSAGE_FAILED_CREATE_SCHEDULE       = "gagal menambahkan schedule"
	MESSAGE_FAILED_UPDATE_SCHEDULE       = "gagal mengupdate schedule"
	MESSAGE_FAILED_DELETE_SCHEDULE       = "gagal menghapus schedule"
)

var (
	ErrAdminCourseNotFound   = myerror.New("course tidak ditemukan", http.StatusNotFound)
	ErrCourseDuplicate       = myerror.New("course dengan code tersebut sudah ada", http.StatusConflict)
	ErrScheduleNotFound      = myerror.New("schedule tidak ditemukan", http.StatusNotFound)
	ErrScheduleDuplicate     = myerror.New("schedule dengan data tersebut sudah ada", http.StatusConflict)
	ErrScheduleFilterMissing = myerror.New("parameter academic_year, term, prodi, dan semester wajib diisi", http.StatusBadRequest)
	ErrInvalidSemester       = myerror.New("semester harus berupa angka", http.StatusBadRequest)
	ErrInvalidUUID           = myerror.New("id tidak valid", http.StatusBadRequest)
	ErrInvalidTimeFormat     = myerror.New("format waktu harus HH:MM", http.StatusBadRequest)
)

type AdminCourseGroupResponse struct {
	Semester    int `json:"semester"`
	TotalCourse int `json:"total_course"`
}

type AdminCourseResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Credit      int     `json:"credit"`
	Semester    int     `json:"semester"`
	IsElective  bool    `json:"is_elective"`
	Description *string `json:"description"`
	Lab         string  `json:"lab"`
}

type AdminCreateCourseRequest struct {
	Code        string  `json:"code" binding:"required" form:"code"`
	Name        string  `json:"name" binding:"required" form:"name"`
	Credit      int     `json:"credit" binding:"required" form:"credit"`
	Semester    int     `json:"semester" binding:"required" form:"semester"`
	IsElective  bool    `json:"is_elective" form:"is_elective"`
	Description *string `json:"description" form:"description"`
	Lab         string  `json:"lab" binding:"required" form:"lab"`
}

type AdminUpdateCourseRequest struct {
	Code        *string `json:"code" form:"code"`
	Name        *string `json:"name" form:"name"`
	Credit      *int    `json:"credit" form:"credit"`
	Semester    *int    `json:"semester" form:"semester"`
	IsElective  *bool   `json:"is_elective" form:"is_elective"`
	Description *string `json:"description" form:"description"`
	Lab         *string `json:"lab" form:"lab"`
}

type AdminUpdateCourseResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Credit      int     `json:"credit"`
	Semester    int     `json:"semester"`
	IsElective  bool    `json:"is_elective"`
	Description *string `json:"description"`
	Lab         string  `json:"lab"`
}

type AdminScheduleGroupResponse struct {
	AcademicYear string `json:"academic_year"`
	Term         string `json:"term"`
	Prodi        string `json:"prodi"`
	Semester     int    `json:"semester"`
}

type AdminScheduleResponse struct {
	ID           string `json:"id"`
	CourseName   string `json:"course_name"`
	LectureName  string `json:"lecture_name"`
	Class        string `json:"class"`
	Day          string `json:"day"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Room         string `json:"room"`
	Semester     int    `json:"semester"`
	AcademicYear string `json:"academic_year"`
	Capacity     int    `json:"capacity"`
	SKS          int    `json:"sks"`
	Prodi        string `json:"prodi"`
	Term         string `json:"term"`
}

type AdminCreateScheduleRequest struct {
	CourseName   string `json:"course_name" binding:"required" form:"course_name"`
	LectureID    string `json:"lecture_id" binding:"required" form:"lecture_id"`
	Class        string `json:"class" binding:"required" form:"class"`
	Day          string `json:"day" binding:"required" form:"day"`
	StartTime    string `json:"start_time" binding:"required" form:"start_time"`
	EndTime      string `json:"end_time" binding:"required" form:"end_time"`
	Room         string `json:"room" binding:"required" form:"room"`
	Semester     int    `json:"semester" binding:"required" form:"semester"`
	AcademicYear string `json:"academic_year" binding:"required" form:"academic_year"`
	Capacity     int    `json:"capacity" binding:"required" form:"capacity"`
	SKS          int    `json:"sks" binding:"required" form:"sks"`
	Prodi        string `json:"prodi" binding:"required" form:"prodi"`
	Term         string `json:"term" binding:"required" form:"term"`
}

type AdminUpdateScheduleRequest struct {
	CourseName   *string `json:"course_name" form:"course_name"`
	LectureID    *string `json:"lecture_id" form:"lecture_id"`
	Class        *string `json:"class" form:"class"`
	Day          *string `json:"day" form:"day"`
	StartTime    *string `json:"start_time" form:"start_time"`
	EndTime      *string `json:"end_time" form:"end_time"`
	Room         *string `json:"room" form:"room"`
	Semester     *int    `json:"semester" form:"semester"`
	AcademicYear *string `json:"academic_year" form:"academic_year"`
	Capacity     *int    `json:"capacity" form:"capacity"`
	SKS          *int    `json:"sks" form:"sks"`
	Prodi        *string `json:"prodi" form:"prodi"`
	Term         *string `json:"term" form:"term"`
}

type AdminUpdateScheduleResponse struct {
	ID           string `json:"id"`
	CourseName   string `json:"course_name"`
	LectureID    string `json:"lecture_id"`
	Class        string `json:"class"`
	Day          string `json:"day"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Room         string `json:"room"`
	Semester     int    `json:"semester"`
	AcademicYear string `json:"academic_year"`
	Capacity     int    `json:"capacity"`
	SKS          int    `json:"sks"`
	Prodi        string `json:"prodi"`
	Term         string `json:"term"`
}

// =========== MESSAGES ===========
const (
	// Lab Path
	MESSAGE_SUCCESS_GET_LAB_PATHS   = "Lab paths retrieved successfully"
	MESSAGE_SUCCESS_CREATE_LAB_PATH = "Lab path created successfully"
	MESSAGE_SUCCESS_UPDATE_LAB_PATH = "Lab path updated successfully"
	MESSAGE_SUCCESS_DELETE_LAB_PATH = "Lab path deleted successfully"
	MESSAGE_FAILED_GET_LAB_PATHS    = "Failed to retrieve lab paths"
	MESSAGE_FAILED_CREATE_LAB_PATH  = "Failed to create lab path"
	MESSAGE_FAILED_UPDATE_LAB_PATH  = "Failed to update lab path"
	MESSAGE_FAILED_DELETE_LAB_PATH  = "Failed to delete lab path"

	// Prerequisite
	MESSAGE_SUCCESS_CREATE_PREREQUISITE = "Prerequisite created successfully"
	MESSAGE_SUCCESS_DELETE_PREREQUISITE = "Prerequisite deleted successfully"
	MESSAGE_FAILED_CREATE_PREREQUISITE  = "Failed to create prerequisite"
	MESSAGE_FAILED_DELETE_PREREQUISITE  = "Failed to delete prerequisite"

	// Path Edge
	MESSAGE_SUCCESS_CREATE_PATH_EDGE = "Path edge created successfully"
	MESSAGE_SUCCESS_DELETE_PATH_EDGE = "Path edge deleted successfully"
	MESSAGE_FAILED_CREATE_PATH_EDGE  = "Failed to create path edge"
	MESSAGE_FAILED_DELETE_PATH_EDGE  = "Failed to delete path edge"

	// Lecture
	MESSAGE_SUCCESS_GET_LECTURES   = "Lectures retrieved successfully"
	MESSAGE_SUCCESS_CREATE_LECTURE = "Lecture created successfully"
	MESSAGE_SUCCESS_UPDATE_LECTURE = "Lecture updated successfully"
	MESSAGE_FAILED_GET_LECTURES    = "Failed to retrieve lectures"
	MESSAGE_FAILED_CREATE_LECTURE  = "Failed to create lecture"
	MESSAGE_FAILED_UPDATE_LECTURE  = "Failed to update lecture"
)

// =========== ERRORS ===========
var (
	ErrLabPathNotFound      = myerror.New("lab path not found", http.StatusNotFound)
	ErrLabPathNameExists    = myerror.New("lab path name already exists", http.StatusConflict)
	ErrAdminCourseNotFound  = myerror.New("course not found", http.StatusNotFound)
	ErrCourseCodeExists     = myerror.New("course code already exists", http.StatusConflict)
	ErrPrerequisiteExists   = myerror.New("prerequisite already exists", http.StatusConflict)
	ErrPrerequisiteNotFound = myerror.New("prerequisite not found", http.StatusNotFound)
	ErrPathEdgeExists       = myerror.New("path edge already exists", http.StatusConflict)
	ErrPathEdgeNotFound     = myerror.New("path edge not found", http.StatusNotFound)
	ErrLectureNotFound      = myerror.New("lecture not found", http.StatusNotFound)
	ErrLectureCodeExists    = myerror.New("lecture code already exists", http.StatusConflict)
)

// =========== LAB PATH ===========

type AdminLabPathResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CreateLabPathRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color" binding:"required"`
}

type UpdateLabPathRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// =========== PREREQUISITE ===========

type AdminPrerequisiteResponse struct {
	ID        string `json:"id"`
	CourseID  string `json:"course_id"`
	RequireID string `json:"require_id"`
}

type CreatePrerequisiteRequest struct {
	CourseID  string `json:"course_id" binding:"required"`
	RequireID string `json:"require_id" binding:"required"`
}

// =========== PATH EDGE ===========

type AdminPathEdgeResponse struct {
	ID           string `json:"id"`
	FromCourseID string `json:"from_course_id"`
	ToCourseID   string `json:"to_course_id"`
}

type CreatePathEdgeRequest struct {
	FromCourseID string `json:"from_course_id" binding:"required"`
	ToCourseID   string `json:"to_course_id" binding:"required"`
}

// =========== LECTURE ===========

type AdminLectureResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateLectureRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type UpdateLectureRequest struct {
	Code *string `json:"code"`
	Name *string `json:"name"`
}
