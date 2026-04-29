package dto

import (
	"net/http"
	"time"

	myerror "ADRIFT-backend/internal/pkg/error"
)

// =========== MESSAGES ===========
const (
	MESSAGE_SUCCESS_GET_GRAPH     = "Graph data retrieved successfully"
	MESSAGE_SUCCESS_GET_NODE      = "Node detail retrieved successfully"
	MESSAGE_SUCCESS_GET_CHAIN     = "Node chain retrieved successfully"
	MESSAGE_SUCCESS_GET_PROGRESS  = "Progress graph retrieved successfully"
	MESSAGE_SUCCESS_GET_SUMMARY   = "Progress summary retrieved successfully"
	MESSAGE_SUCCESS_CLAIM_COURSE  = "Course claimed successfully"
	MESSAGE_SUCCESS_UNCLAIM_COURSE = "Course unclaimed successfully"

	MESSAGE_FAILED_GET_GRAPH    = "Failed to retrieve graph data"
	MESSAGE_FAILED_GET_NODE     = "Failed to retrieve node detail"
	MESSAGE_FAILED_GET_CHAIN    = "Failed to retrieve node chain"
	MESSAGE_FAILED_GET_PROGRESS = "Failed to retrieve progress graph"
	MESSAGE_FAILED_GET_SUMMARY  = "Failed to retrieve progress summary"
	MESSAGE_FAILED_CLAIM_COURSE = "Failed to claim course"
	MESSAGE_FAILED_UNCLAIM      = "Failed to unclaim course"
)

// =========== ERRORS ===========
var (
	ErrCourseNotFound     = myerror.New("course not found", http.StatusNotFound)
	ErrCourseNotAvailable = myerror.New("course is not available to claim", http.StatusBadRequest)
	ErrCourseNotCompleted = myerror.New("course is not completed", http.StatusBadRequest)
	ErrProgressNotFound   = myerror.New("progress record not found", http.StatusNotFound)
)

// =========== SHARED ===========

type LabPathResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// =========== PUBLIC GRAPH ===========

type GraphNodeResponse struct {
	ID          string            `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Credit      int               `json:"credit"`
	Semester    int               `json:"semester"`
	IsElective  bool              `json:"is_elective"`
	Description *string           `json:"description"`
	LabPaths    []LabPathResponse `json:"lab_paths"`
}

type GraphEdgeResponse struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`  // "PREREQUISITE" | "PATH"
	Color  string `json:"color"`
}

type GraphResponse struct {
	Nodes []GraphNodeResponse `json:"nodes"`
	Edges []GraphEdgeResponse `json:"edges"`
}

// =========== NODE DETAIL ===========

type NodeScheduleResponse struct {
	ID          string `json:"id"`
	Class       string `json:"class"`
	Day         string `json:"day"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	LectureName string `json:"lecture_name"`
	Room        string `json:"room"`
	Capacity    int    `json:"capacity"`
}

type NodeCourseRef struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Credit     int    `json:"credit"`
	Semester   int    `json:"semester"`
	IsElective bool   `json:"is_elective,omitempty"`
}

type NodeDetailResponse struct {
	ID                    string                 `json:"id"`
	Code                  string                 `json:"code"`
	Name                  string                 `json:"name"`
	Credit                int                    `json:"credit"`
	Semester              int                    `json:"semester"`
	IsElective            bool                   `json:"is_elective"`
	Description           *string                `json:"description"`
	LabPaths              []LabPathResponse      `json:"lab_paths"`
	Prerequisites         []NodeCourseRef        `json:"prerequisites"`
	Unlocks               []NodeCourseRef        `json:"unlocks"`
	SchedulesThisSemester []NodeScheduleResponse `json:"schedules_this_semester"`
}

// =========== NODE CHAIN ===========

type NodeChainResponse struct {
	Upstream   []string `json:"upstream"`
	Downstream []string `json:"downstream"`
}

// =========== PROGRESS GRAPH ===========

type ProgressNodeResponse struct {
	ID          string            `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Credit      int               `json:"credit"`
	Semester    int               `json:"semester"`
	IsElective  bool              `json:"is_elective"`
	Description *string           `json:"description"`
	Status      string            `json:"status"` // COMPLETED | AVAILABLE | LOCKED
	Grade       *string           `json:"grade"`
	ClaimedAt   *time.Time        `json:"claimed_at"`
	LabPaths    []LabPathResponse `json:"lab_paths"`
}

type ProgressGraphResponse struct {
	Nodes []ProgressNodeResponse `json:"nodes"`
	Edges []GraphEdgeResponse    `json:"edges"`
}

// =========== CLAIM ===========

type ClaimCourseRequest struct {
	Grade *string `json:"grade"`
}

type StatusChangeItem struct {
	CourseID   string `json:"course_id"`
	CourseCode string `json:"course_code"`
	CourseName string `json:"course_name"`
	Status     string `json:"status"`
}

type ClaimCourseResponse struct {
	NewlyAvailable []StatusChangeItem `json:"newly_available"`
}

type UnclaimCourseResponse struct {
	NewlyLocked []StatusChangeItem `json:"newly_locked"`
}

// =========== SUMMARY ===========

type ProgressSummaryResponse struct {
	TotalCourses            int     `json:"total_courses"`
	Completed               int     `json:"completed"`
	Available               int     `json:"available"`
	Locked                  int     `json:"locked"`
	TotalCreditsCompleted   int     `json:"total_credits_completed"`
	TotalCreditsRequired    int     `json:"total_credits_required"`
	CurrentSemesterEstimate int     `json:"current_semester_estimate"`
	EnrollmentYear          int     `json:"enrollment_year"`
	ProgressPercentage      float64 `json:"progress_percentage"`
}
