package entity

import "github.com/google/uuid"

type LabPath struct {
	ID    uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`

	PathEdges []PathEdge `gorm:"foreignKey:LabPathID;references:ID" json:"path_edges,omitempty"`

	Timestamp
}

func (lp *LabPath) TableName() string {
	return "lab_paths"
}
