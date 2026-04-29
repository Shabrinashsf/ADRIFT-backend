package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FRSPlanItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	FRSPlanID  uuid.UUID `json:"frs_plan_id"`
	ScheduleID uuid.UUID `json:"schedule_id"`

	FRSPlan  *FRSPlan  `gorm:"foreignKey:FRSPlanID;references:ID"`
	Schedule *Schedule `gorm:"foreignKey:ScheduleID;references:ID"`
}

func (fpi *FRSPlanItem) BeforeCreate(tx *gorm.DB) error {
	if fpi.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		fpi.ID = newID
	}
	return nil
}

func (fpi *FRSPlanItem) TableName() string {
	return "frs_plan_items"
}
