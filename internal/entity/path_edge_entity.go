package entity

import "github.com/google/uuid"

type PathEdge struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	FromCourseID uuid.UUID `json:"from_course_id"`
	ToCourseID   uuid.UUID `json:"to_course_id"`

	FromCourse *Course `gorm:"foreignKey:FromCourseID;references:ID"`
	ToCourse   *Course `gorm:"foreignKey:ToCourseID;references:ID"`
}

func (pe *PathEdge) TableName() string {
	return "path_edges"
}
