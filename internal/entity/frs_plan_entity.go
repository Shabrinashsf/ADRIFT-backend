package entity

import "github.com/google/uuid"

type FRSPlan struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	PlanNamwe    string       `json:"plan_name"`
	AcademicYear string       `json:"academic_year"`
	Term         TermSemester `json:"term"`
	TotalCredit  int          `json:"total_credit"`

	FRSPlanItems []FRSPlanItem `gorm:"foreignKey:FRSPlanID;references:ID" json:"frs_plan_items,omitempty"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

func (fp *FRSPlan) TableName() string {
	return "frs_plans"
}
