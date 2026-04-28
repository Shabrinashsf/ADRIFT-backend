package entity

import "github.com/google/uuid"

type FRSPlanItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	FRSPlanID  uuid.UUID `json:"frs_plan_id"`
	ScheduleID uuid.UUID `json:"schedule_id"`

	FRSPlan  *FRSPlan  `gorm:"foreignKey:FRSPlanID;references:ID"`
	Schedule *Schedule `gorm:"foreignKey:ScheduleID;references:ID"`
}

func (fpi *FRSPlanItem) TableName() string {
	return "frs_plan_items"
}
