package entity

import "github.com/google/uuid"

type FRSPlanItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	FRSPlanID  uuid.UUID `json:"frs_plan_id"`
	ScheduleID uuid.UUID `json:"schedule_id"`

	FRSPlan  *FRSPlan  `gorm:"foreignKey:FRSPlanID"`
	Schedule *Schedule `gorm:"foreignKey:ScheduleID"`
}

func (fpi *FRSPlanItem) TableName() string {
	return "frs_plan_items"
}
