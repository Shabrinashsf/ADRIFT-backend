package dto

import (
	"net/http"

	myerror "ADRIFT-backend/internal/pkg/error"
)

// =========== MESSAGES ===========
const (
	// Course
	MESSAGE_SUCCESS_GET_COURSES    = "Courses retrieved successfully"
	MESSAGE_SUCCESS_CREATE_COURSE  = "Course created successfully"
	MESSAGE_SUCCESS_UPDATE_COURSE  = "Course updated successfully"
	MESSAGE_SUCCESS_DELETE_COURSE  = "Course deleted successfully"
	MESSAGE_FAILED_GET_COURSES     = "Failed to retrieve courses"
	MESSAGE_FAILED_CREATE_COURSE   = "Failed to create course"
	MESSAGE_FAILED_UPDATE_COURSE   = "Failed to update course"
	MESSAGE_FAILED_DELETE_COURSE   = "Failed to delete course"

	// Lab Path
	MESSAGE_SUCCESS_GET_LAB_PATHS    = "Lab paths retrieved successfully"
	MESSAGE_SUCCESS_CREATE_LAB_PATH  = "Lab path created successfully"
	MESSAGE_SUCCESS_UPDATE_LAB_PATH  = "Lab path updated successfully"
	MESSAGE_SUCCESS_DELETE_LAB_PATH  = "Lab path deleted successfully"
	MESSAGE_FAILED_GET_LAB_PATHS     = "Failed to retrieve lab paths"
	MESSAGE_FAILED_CREATE_LAB_PATH   = "Failed to create lab path"
	MESSAGE_FAILED_UPDATE_LAB_PATH   = "Failed to update lab path"
	MESSAGE_FAILED_DELETE_LAB_PATH   = "Failed to delete lab path"

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
	ErrLabPathNotFound     = myerror.New("lab path not found", http.StatusNotFound)
	ErrLabPathNameExists   = myerror.New("lab path name already exists", http.StatusConflict)
	ErrAdminCourseNotFound = myerror.New("course not found", http.StatusNotFound)
	ErrCourseCodeExists    = myerror.New("course code already exists", http.StatusConflict)
	ErrPrerequisiteExists  = myerror.New("prerequisite already exists", http.StatusConflict)
	ErrPrerequisiteNotFound = myerror.New("prerequisite not found", http.StatusNotFound)
	ErrPathEdgeExists      = myerror.New("path edge already exists", http.StatusConflict)
	ErrPathEdgeNotFound    = myerror.New("path edge not found", http.StatusNotFound)
	ErrLectureNotFound     = myerror.New("lecture not found", http.StatusNotFound)
	ErrLectureCodeExists   = myerror.New("lecture code already exists", http.StatusConflict)
)

// =========== COURSE ===========

type AdminCourseResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Credit      int     `json:"credit"`
	Semester    int     `json:"semester"`
	IsElective  bool    `json:"is_elective"`
	Lab         string  `json:"lab"`
	Description *string `json:"description"`
}

type CreateCourseRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Credit      int     `json:"credit" binding:"required"`
	Semester    int     `json:"semester" binding:"required"`
	IsElective  bool    `json:"is_elective"`
	Lab         string  `json:"lab" binding:"required"`
	Description *string `json:"description"`
}

type UpdateCourseRequest struct {
	Code        *string `json:"code"`
	Name        *string `json:"name"`
	Credit      *int    `json:"credit"`
	Semester    *int    `json:"semester"`
	IsElective  *bool   `json:"is_elective"`
	Lab         *string `json:"lab"`
	Description *string `json:"description"`
}

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
