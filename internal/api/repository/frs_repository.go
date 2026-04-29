package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"ADRIFT-backend/internal/entity"
	"ADRIFT-backend/internal/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	FRSRepository interface {
		ScheduleExists(ctx context.Context, academicYear string, term entity.TermSemester) (bool, error)
		CreateSchedules(ctx context.Context, schedules []entity.Schedule) error
		ListLectures(ctx context.Context) ([]entity.Lecture, error)
		EnsureLectureByCode(ctx context.Context, code string) (uuid.UUID, error)
		ListSchedules(ctx context.Context, tx *gorm.DB, meta pagination.Meta) ([]entity.Schedule, int, error)
		CreateFRSPlan(ctx context.Context, tx *gorm.DB, plan entity.FRSPlan) (entity.FRSPlan, error)
		CreateFRSPlanItems(ctx context.Context, tx *gorm.DB, items []entity.FRSPlanItem) error
		GetFRSPlanByID(ctx context.Context, planID uuid.UUID, userID uuid.UUID) (entity.FRSPlan, error)
		ListFRSPlansByUserID(ctx context.Context, userID uuid.UUID) ([]entity.FRSPlan, error)
		FindSchedulesByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Schedule, error)
		FindSchedulesByLectureIDsAndTerm(ctx context.Context, lectureIDs []uuid.UUID, academicYear string, term entity.TermSemester) ([]entity.Schedule, error)
		FindSchedulesByCourseNamesAndTerm(ctx context.Context, courseNames []string, academicYear string, term entity.TermSemester) ([]entity.Schedule, error)
		DeleteFRSPlan(ctx context.Context, planID uuid.UUID, userID uuid.UUID) error
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

func (r *frsRepository) ListSchedules(ctx context.Context, tx *gorm.DB, meta pagination.Meta) ([]entity.Schedule, int, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx).Model(&entity.Schedule{})

	allowedILikeFilters := map[string]bool{
		"academic_year": true,
		"prodi":         true,
		"course_name":   true,
	}

	for key, val := range meta.Filters {
		if key == "semester" {
			semester, err := strconv.Atoi(strings.TrimSpace(val))
			if err != nil {
				query = query.Where("1 = 0")
			} else {
				query = query.Where("schedules.semester = ?", semester)
			}
		} else if key == "term" {
			query = query.Where("schedules.term = ?", val)
		} else if allowedILikeFilters[key] {
			query = query.Where("schedules."+key+" ILIKE ?", "%"+val+"%")
		}
	}

	var totalData int64
	if err := query.Count(&totalData).Error; err != nil {
		return nil, 0, err
	}

	skip, limit := meta.GetSkipAndLimit()

	sortBy := strings.ToLower(meta.SortBy)
	allowedSortBy := map[string]bool{
		"course_name":   true,
		"class":         true,
		"day":           true,
		"room":          true,
		"capacity":      true,
		"prodi":         true,
		"academic_year": true,
		"term":          true,
		"semester":      true,
	}
	if !allowedSortBy[sortBy] {
		sortBy = "course_name"
	}

	sort := strings.ToLower(meta.Sort)
	if sort != "desc" {
		sort = "asc"
	}

	var schedules []entity.Schedule
	err := query.
		Preload("Lecture").
		Order("schedules." + sortBy + " " + sort).
		Offset(skip).
		Limit(limit).
		Find(&schedules).Error
	if err != nil {
		return nil, 0, err
	}

	return schedules, int(totalData), nil
}

func (r *frsRepository) CreateFRSPlan(ctx context.Context, tx *gorm.DB, plan entity.FRSPlan) (entity.FRSPlan, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Create(&plan).Error; err != nil {
		return entity.FRSPlan{}, err
	}
	return plan, nil
}

func (r *frsRepository) CreateFRSPlanItems(ctx context.Context, tx *gorm.DB, items []entity.FRSPlanItem) error {
	if tx == nil {
		tx = r.db
	}
	if len(items) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&items).Error
}

func (r *frsRepository) GetFRSPlanByID(ctx context.Context, planID uuid.UUID, userID uuid.UUID) (entity.FRSPlan, error) {
	var plan entity.FRSPlan
	if err := r.db.WithContext(ctx).
		Preload("FRSPlanItems.Schedule.Lecture").
		Where("id = ? AND user_id = ?", planID, userID).
		First(&plan).Error; err != nil {
		return entity.FRSPlan{}, err
	}
	return plan, nil
}

func (r *frsRepository) ListFRSPlansByUserID(ctx context.Context, userID uuid.UUID) ([]entity.FRSPlan, error) {
	var plans []entity.FRSPlan
	if err := r.db.WithContext(ctx).
		Preload("FRSPlanItems").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *frsRepository) FindSchedulesByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Schedule, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var schedules []entity.Schedule
	if err := r.db.WithContext(ctx).
		Preload("Lecture").
		Where("id IN ?", ids).
		Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *frsRepository) FindSchedulesByLectureIDsAndTerm(ctx context.Context, lectureIDs []uuid.UUID, academicYear string, term entity.TermSemester) ([]entity.Schedule, error) {
	if len(lectureIDs) == 0 {
		return nil, nil
	}
	var schedules []entity.Schedule
	if err := r.db.WithContext(ctx).
		Preload("Lecture").
		Where("lecture_id IN ? AND academic_year = ? AND term = ?", lectureIDs, academicYear, term).
		Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *frsRepository) FindSchedulesByCourseNamesAndTerm(ctx context.Context, courseNames []string, academicYear string, term entity.TermSemester) ([]entity.Schedule, error) {
	if len(courseNames) == 0 {
		return nil, nil
	}
	var schedules []entity.Schedule
	if err := r.db.WithContext(ctx).
		Preload("Lecture").
		Where("course_name IN ? AND academic_year = ? AND term = ?", courseNames, academicYear, term).
		Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *frsRepository) DeleteFRSPlan(ctx context.Context, planID uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", planID, userID).
		Delete(&entity.FRSPlan{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
