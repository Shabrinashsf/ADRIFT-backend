package entity

import "github.com/google/uuid"

type LabPath struct {
	ID    uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name  string    `gorm:"type:text;uniqueIndex" json:"name"`
	Color string    `gorm:"type:text" json:"color"`

	PathEdges     []PathEdge     `gorm:"foreignKey:Lab;references:Name" json:"path_edges,omitempty"`
	Prerequisites []Prerequisite `gorm:"foreignKey:Lab;references:Name" json:"prerequisites,omitempty"`

	Timestamp
}

func (lp *LabPath) TableName() string {
	return "lab_paths"
}
