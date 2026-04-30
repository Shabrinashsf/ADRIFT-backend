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
