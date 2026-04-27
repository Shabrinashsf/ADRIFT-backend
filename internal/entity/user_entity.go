package entity

import (
	"ADRIFT-backend/internal/pkg/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleStudent UserRole = "STUDENT"
	UserRoleAdmin   UserRole = "ADMIN"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	NRP            string    `json:"nrp" gorm:"uniqueIndex"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	EnrollmentYear int       `json:"enrollment_year"`
	Role           UserRole  `json:"role" gorm:"default:STUDENT"`
	IsVerified     bool      `json:"is_verified"`

	FRSPlans        []FRSPlan         `json:"frs_plans,omitempty"`
	StudentProgress []StudentProgress `json:"student_progress,omitempty"`

	Timestamp
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		newID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.ID = newID
	}
	return nil
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := utils.HashPassword(u.Password)
		if err != nil {
			return err
		}
		u.Password = hashedPassword
	}
	return nil
}

func (u *User) TableName() string {
	return "users"
}
