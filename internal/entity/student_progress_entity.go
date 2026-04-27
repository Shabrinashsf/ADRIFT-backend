package entity

import (
	"time"

	"github.com/google/uuid"
)

type NodeStatus string

const (
	NodeStatusAvailable NodeStatus = "AVAILABLE"
	NodeStatusLocked    NodeStatus = "LOCKED"
	NodeStatusDisable   NodeStatus = "COMPLETED"
)

type StudentProgress struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;" json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	CourseID  uuid.UUID  `json:"course_id"`
	Status    NodeStatus `json:"status"`
	Grade     *string    `json:"grade"`
	ClaimedAt *time.Time `json:"claimed_at"`

	User   *User   `gorm:"foreignKey:UserID"`
	Course *Course `gorm:"foreignKey:CourseID"`
}

func (sp *StudentProgress) TableName() string {
	return "student_progress"
}
