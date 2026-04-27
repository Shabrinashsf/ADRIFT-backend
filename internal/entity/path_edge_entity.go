package entity

import "github.com/google/uuid"

type PathEdge struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	FromCourseID uuid.UUID `json:"from_course_id"`
	ToCourseID   uuid.UUID `json:"to_course_id"`
	LabPathID    uuid.UUID `json:"lab_path_id"`

	FromCourse *Course  `gorm:"foreignKey:FromCourseID"`
	ToCourse   *Course  `gorm:"foreignKey:ToCourseID"`
	LabPath    *LabPath `gorm:"foreignKey:LabPathID"`
}

func (pe *PathEdge) TableName() string {
	return "path_edges"
}
