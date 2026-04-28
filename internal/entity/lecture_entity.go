package entity

import "github.com/google/uuid"

type Lecture struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`

	Schedules []Schedule `gorm:"foreignKey:LectureID;references:ID" json:"schedules,omitempty"`
}

func (l *Lecture) TableName() string {
	return "lectures"
}
