package repository

import "gorm.io/gorm"

type (
	AdminRepository interface {
	}

	adminRepository struct {
		db *gorm.DB
	}
)

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{
		db: db,
	}
}
