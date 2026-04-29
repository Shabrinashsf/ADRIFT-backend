package entity

import "github.com/google/uuid"

type Prerequisite struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	RequireID uuid.UUID `json:"require_id"`
	Lab       string    `gorm:"type:text" json:"lab"`

	Course  *Course  `gorm:"foreignKey:CourseID;references:ID"`
	Require *Course  `gorm:"foreignKey:RequireID;references:ID"`
	LabPath *LabPath `gorm:"foreignKey:Lab;references:Name"`
}

func (p *Prerequisite) TableName() string {
	return "prerequisites"
}
