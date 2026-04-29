package repository

import (
	"context"

	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	SkillTreeRepository interface {
		GetAllCourses(ctx context.Context) ([]entity.Course, error)
		GetAllPrerequisites(ctx context.Context) ([]entity.Prerequisite, error)
		GetAllPathEdges(ctx context.Context) ([]entity.PathEdge, error)
		GetCourseByID(ctx context.Context, courseID uuid.UUID) (*entity.Course, error)
		GetSchedulesByCourseID(ctx context.Context, courseID uuid.UUID) ([]entity.Schedule, error)
		GetUnlocksByCourseID(ctx context.Context, courseID uuid.UUID) ([]entity.Course, error)
		GetProgressByUserID(ctx context.Context, userID uuid.UUID) ([]entity.StudentProgress, error)
		GetProgressByCourseAndUser(ctx context.Context, userID, courseID uuid.UUID) (*entity.StudentProgress, error)
		UpsertProgress(ctx context.Context, progress *entity.StudentProgress) error
		DeleteProgress(ctx context.Context, userID, courseID uuid.UUID) error
		CountAllCourses(ctx context.Context) (int64, error)
		GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)
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

func (r *skillTreeRepository) GetAllCourses(ctx context.Context) ([]entity.Course, error) {
	var courses []entity.Course
	err := r.db.WithContext(ctx).
		Preload("LabPath").
		Find(&courses).Error
	return courses, err
}

func (r *skillTreeRepository) GetAllPrerequisites(ctx context.Context) ([]entity.Prerequisite, error) {
	var prereqs []entity.Prerequisite
	err := r.db.WithContext(ctx).Find(&prereqs).Error
	return prereqs, err
}

func (r *skillTreeRepository) GetAllPathEdges(ctx context.Context) ([]entity.PathEdge, error) {
	var edges []entity.PathEdge
	err := r.db.WithContext(ctx).Find(&edges).Error
	return edges, err
}

func (r *skillTreeRepository) GetCourseByID(ctx context.Context, courseID uuid.UUID) (*entity.Course, error) {
	var course entity.Course
	err := r.db.WithContext(ctx).
		Preload("LabPath").
		Preload("PrerequisiteCourses.Require").
		Where("id = ?", courseID).
		First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *skillTreeRepository) GetSchedulesByCourseID(ctx context.Context, courseID uuid.UUID) ([]entity.Schedule, error) {
	var schedules []entity.Schedule
	err := r.db.WithContext(ctx).
		Preload("Lecture").
		Where("course_id = ?", courseID).
		Find(&schedules).Error
	return schedules, err
}

// GetUnlocksByCourseID: returns courses that have courseID as a prerequisite OR as a path edge source
func (r *skillTreeRepository) GetUnlocksByCourseID(ctx context.Context, courseID uuid.UUID) ([]entity.Course, error) {
	// Courses that list courseID as prerequisite
	var prereqUnlocks []entity.Course
	r.db.WithContext(ctx).
		Joins("JOIN prerequisites ON prerequisites.course_id = courses.id").
		Where("prerequisites.require_id = ?", courseID).
		Find(&prereqUnlocks)

	// Courses reachable via path_edges
	var pathUnlocks []entity.Course
	r.db.WithContext(ctx).
		Joins("JOIN path_edges ON path_edges.to_course_id = courses.id").
		Where("path_edges.from_course_id = ?", courseID).
		Find(&pathUnlocks)

	// Merge, deduplicate
	seen := map[uuid.UUID]bool{}
	result := []entity.Course{}
	for _, c := range append(prereqUnlocks, pathUnlocks...) {
		if !seen[c.ID] {
			seen[c.ID] = true
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *skillTreeRepository) GetProgressByUserID(ctx context.Context, userID uuid.UUID) ([]entity.StudentProgress, error) {
	var progress []entity.StudentProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&progress).Error
	return progress, err
}

func (r *skillTreeRepository) GetProgressByCourseAndUser(ctx context.Context, userID, courseID uuid.UUID) (*entity.StudentProgress, error) {
	var progress entity.StudentProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *skillTreeRepository) UpsertProgress(ctx context.Context, progress *entity.StudentProgress) error {
	return r.db.WithContext(ctx).Save(progress).Error
}

func (r *skillTreeRepository) DeleteProgress(ctx context.Context, userID, courseID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Delete(&entity.StudentProgress{}).Error
}

func (r *skillTreeRepository) CountAllCourses(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Course{}).Count(&count).Error
	return count, err
}

func (r *skillTreeRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
