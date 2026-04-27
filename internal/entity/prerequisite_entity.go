package entity

import "github.com/google/uuid"

type Prerequisite struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	RequireID uuid.UUID `json:"require_id"`

	Course  *Course `gorm:"foreignKey:CourseID"`
	Require *Course `gorm:"foreignKey:RequireID"`
}

func (p *Prerequisite) TableName() string {
	return "prerequisites"
}
