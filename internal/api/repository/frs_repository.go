package repository

import (
	"context"
	"errors"
	"strings"

	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	FRSRepository interface {
		ScheduleExists(ctx context.Context, academicYear string, term entity.TermSemester) (bool, error)
		CreateSchedules(ctx context.Context, schedules []entity.Schedule) error
		ListLectures(ctx context.Context) ([]entity.Lecture, error)
		EnsureLectureByCode(ctx context.Context, code string) (uuid.UUID, error)
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

func (r *frsRepository) ListLectures(ctx context.Context) ([]entity.Lecture, error) {
	var lectures []entity.Lecture
	if err := r.db.WithContext(ctx).Find(&lectures).Error; err != nil {
		return nil, err
	}

	return lectures, nil
}

func (r *frsRepository) EnsureLectureByCode(ctx context.Context, code string) (uuid.UUID, error) {
	cleaned := strings.TrimSpace(code)
	if cleaned == "" {
		return uuid.Nil, nil
	}

	var lecture entity.Lecture
	err := r.db.WithContext(ctx).Where(&entity.Lecture{Code: cleaned}).First(&lecture).Error
	if err == nil {
		return lecture.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	lecture = entity.Lecture{
		ID:   uuid.New(),
		Code: cleaned,
		Name: cleaned,
	}
	if err := r.db.WithContext(ctx).Create(&lecture).Error; err != nil {
		return uuid.Nil, err
	}

	return lecture.ID, nil
}
