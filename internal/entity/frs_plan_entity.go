package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FRSPlan struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	PlanName     string       `gorm:"column:plan_namwe" json:"plan_name"`
	AcademicYear string       `json:"academic_year"`
	Term         TermSemester `json:"term"`
	TotalCredit  int          `json:"total_credit"`

	FRSPlanItems []FRSPlanItem `gorm:"foreignKey:FRSPlanID;References:ID" json:"frs_plan_items,omitempty"`

	User *User `gorm:"foreignKey:UserID;references:ID"`

	Timestamp
}

func (fp *FRSPlan) BeforeCreate(tx *gorm.DB) error {
	if fp.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		fp.ID = newID
	}
	return nil
}

func (fp *FRSPlan) TableName() string {
	return "frs_plans"
}
