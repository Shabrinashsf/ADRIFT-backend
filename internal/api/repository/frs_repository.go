package repository

import "gorm.io/gorm"

type (
	FRSRepository interface {
	}

	frsRepository struct {
		db *gorm.DB
	}
)

func NewFRSRepository(db *gorm.DB) FRSRepository {
	return &frsRepository{
		db: db,
	}
}
