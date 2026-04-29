package repository

import (
	"context"

	"ADRIFT-backend/internal/entity"

	"gorm.io/gorm"
)

type (
	FRSRepository interface {
		ScheduleExists(ctx context.Context, academicYear string, term entity.TermSemester) (bool, error)
		CreateSchedules(ctx context.Context, schedules []entity.Schedule) error
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

func (r *frsRepository) ScheduleExists(ctx context.Context, academicYear string, term entity.TermSemester) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Schedule{}).
		Where("academic_year = ? AND term = ?", academicYear, term).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *frsRepository) CreateSchedules(ctx context.Context, schedules []entity.Schedule) error {
	if len(schedules) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Create(&schedules).Error
}
