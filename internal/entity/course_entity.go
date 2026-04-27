package entity

import "github.com/google/uuid"

type Course struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Credit      int       `json:"credit"`
	Semester    int       `json:"semester"`
	IsElective  bool      `json:"is_elective"`
	Description *string   `json:"description"`

	Schedules                []Schedule        `json:"schedules,omitempty"`
	PathEdgesFrom            []PathEdge        `json:"path_edges_from,omitempty"`            // ini yang from_course_id di tabel path_edge
	PathEdgesTo              []PathEdge        `json:"path_edges_to,omitempty"`              // ini yang to_course_id di tabel path_edge
	PrerequisitesRequiredFor []Prerequisite    `json:"prerequisites_required_for,omitempty"` // ini yang course_id di tabel prerequisite
	PrerequisiteCourses      []Prerequisite    `json:"prerequisite_courses,omitempty"`       // ini yang require_id di tabel prerequisite
	StudentProgress          []StudentProgress `json:"student_progress,omitempty"`

	Timestamp
}

func (c *Course) TableName() string {
	return "courses"
}
