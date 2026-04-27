package repository

import "gorm.io/gorm"

type (
	SkillTreeRepository interface {
	}

	skillTreeRepository struct {
		db *gorm.DB
	}
)

func NewSkillTreeRepository(db *gorm.DB) SkillTreeRepository {
	return &skillTreeRepository{
		db: db,
	}
}
