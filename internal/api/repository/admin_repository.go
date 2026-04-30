package repository

import (
	"context"

	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourseGroupResult struct {
	Semester    int
	TotalCourse int
}

type ScheduleGroupResult struct {
	AcademicYear string
	Term         string
	Prodi        string
	Semester     int
}

type (
	AdminRepository interface {
		GetCoursesGroupedBySemester(ctx context.Context) ([]CourseGroupResult, error)
GetCoursesBySemester(ctx context.Context, semester int, courseName string) ([]entity.Course, error)
		GetCourseByCode(ctx context.Context, code string) (entity.Course, error)
		GetCourseByID(ctx context.Context, id uuid.UUID) (entity.Course, error)
		CreateCourse(ctx context.Context, course entity.Course) (entity.Course, error)
		UpdateCourse(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (entity.Course, error)
		SoftDeleteCourse(ctx context.Context, id uuid.UUID) error
		GetScheduleGroups(ctx context.Context) ([]ScheduleGroupResult, error)
GetSchedulesByFilter(ctx context.Context, academicYear, term string, prodi entity.ProdiType, semester int, courseName string) ([]entity.Schedule, error)
		GetScheduleByID(ctx context.Context, id uuid.UUID) (entity.Schedule, error)
		CheckScheduleDuplicate(ctx context.Context, courseName, class, academicYear, term string, prodi entity.ProdiType, semester int) (bool, error)
		CheckScheduleDuplicateExcludeID(ctx context.Context, excludeID uuid.UUID, courseName, class, academicYear, term string, prodi entity.ProdiType, semester int) (bool, error)
		CreateSchedule(ctx context.Context, schedule entity.Schedule) (entity.Schedule, error)
		UpdateSchedule(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (entity.Schedule, error)
DeleteSchedule(ctx context.Context, id uuid.UUID) error
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

func (r *adminRepository) GetCoursesGroupedBySemester(ctx context.Context) ([]CourseGroupResult, error) {
	var groups []CourseGroupResult
	err := r.db.WithContext(ctx).
		Model(&entity.Course{}).
		Select("semester, count(*) as total_course").
		Where("deleted_at IS NULL").
		Group("semester").
		Order("semester asc").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *adminRepository) GetCoursesBySemester(ctx context.Context, semester int, courseName string) ([]entity.Course, error) {
	var courses []entity.Course
	q := r.db.WithContext(ctx).
		Where("semester = ? AND deleted_at IS NULL", semester)
	if courseName != "" {
		q = q.Where("name ILIKE ?", "%"+courseName+"%")
	}
	err := q.Order("name asc").Find(&courses).Error
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func (r *adminRepository) GetCourseByCode(ctx context.Context, code string) (entity.Course, error) {
	var course entity.Course
	err := r.db.WithContext(ctx).
		Where("code = ? AND deleted_at IS NULL", code).
		First(&course).Error
	return course, err
}

func (r *adminRepository) GetCourseByID(ctx context.Context, id uuid.UUID) (entity.Course, error) {
	var course entity.Course
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&course).Error
	return course, err
}

func (r *adminRepository) CreateCourse(ctx context.Context, course entity.Course) (entity.Course, error) {
	err := r.db.WithContext(ctx).Create(&course).Error
	return course, err
}

func (r *adminRepository) UpdateCourse(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (entity.Course, error) {
	err := r.db.WithContext(ctx).
		Model(&entity.Course{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates).Error
	if err != nil {
		return entity.Course{}, err
	}
	return r.GetCourseByID(ctx, id)
}

func (r *adminRepository) SoftDeleteCourse(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&entity.Course{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) GetScheduleGroups(ctx context.Context) ([]ScheduleGroupResult, error) {
	var groups []ScheduleGroupResult
	err := r.db.WithContext(ctx).
		Model(&entity.Schedule{}).
		Select("DISTINCT academic_year, term, prodi, semester").
		Where("deleted_at IS NULL").
		Order("academic_year desc, term, prodi, semester").
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *adminRepository) GetSchedulesByFilter(ctx context.Context, academicYear, term string, prodi entity.ProdiType, semester int, courseName string) ([]entity.Schedule, error) {
	var schedules []entity.Schedule
	q := r.db.WithContext(ctx).
		Preload("Lecture").
		Where("academic_year = ? AND term = ? AND prodi = ? AND semester = ? AND deleted_at IS NULL", academicYear, term, prodi, semester)
	if courseName != "" {
		q = q.Where("course_name ILIKE ?", "%"+courseName+"%")
	}
	err := q.Order("course_name asc").Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *adminRepository) GetScheduleByID(ctx context.Context, id uuid.UUID) (entity.Schedule, error) {
	var schedule entity.Schedule
	err := r.db.WithContext(ctx).
		Preload("Lecture").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&schedule).Error
	return schedule, err
}

func (r *adminRepository) CheckScheduleDuplicate(ctx context.Context, courseName, class, academicYear, term string, prodi entity.ProdiType, semester int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Schedule{}).
		Where("course_name = ? AND class = ? AND academic_year = ? AND term = ? AND prodi = ? AND semester = ? AND deleted_at IS NULL",
			courseName, class, academicYear, term, prodi, semester).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *adminRepository) CheckScheduleDuplicateExcludeID(ctx context.Context, excludeID uuid.UUID, courseName, class, academicYear, term string, prodi entity.ProdiType, semester int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Schedule{}).
		Where("course_name = ? AND class = ? AND academic_year = ? AND term = ? AND prodi = ? AND semester = ? AND id != ? AND deleted_at IS NULL",
			courseName, class, academicYear, term, prodi, semester, excludeID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *adminRepository) CreateSchedule(ctx context.Context, schedule entity.Schedule) (entity.Schedule, error) {
	err := r.db.WithContext(ctx).Create(&schedule).Error
	if err != nil {
		return entity.Schedule{}, err
	}
	return r.GetScheduleByID(ctx, schedule.ID)
}

func (r *adminRepository) UpdateSchedule(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (entity.Schedule, error) {
	err := r.db.WithContext(ctx).
		Model(&entity.Schedule{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates).Error
	if err != nil {
		return entity.Schedule{}, err
	}
	return r.GetScheduleByID(ctx, id)
}

func (r *adminRepository) DeleteSchedule(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&entity.Schedule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}