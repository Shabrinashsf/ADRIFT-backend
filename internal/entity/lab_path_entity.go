package entity

import "github.com/google/uuid"

type LabPath struct {
	ID    uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name  string    `gorm:"type:text;uniqueIndex" json:"name"`
	Color string    `gorm:"type:text" json:"color"`

	Courses []Course `gorm:"foreignKey:Lab;references:Name" json:"courses,omitempty"`

	Timestamp
}

func (lp *LabPath) TableName() string {
	return "lab_paths"
}
