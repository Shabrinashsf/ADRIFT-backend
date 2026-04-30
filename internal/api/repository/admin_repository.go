package repository

import (
	"context"

	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AdminRepository interface {
		// Course
		GetAllCourses(ctx context.Context) ([]entity.Course, error)
		GetCourseByID(ctx context.Context, id uuid.UUID) (*entity.Course, error)
		GetCourseByCode(ctx context.Context, code string) (*entity.Course, error)
		CreateCourse(ctx context.Context, course *entity.Course) error
		UpdateCourse(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*entity.Course, error)
		DeleteCourse(ctx context.Context, id uuid.UUID) error

		// Lab Path
		GetAllLabPaths(ctx context.Context) ([]entity.LabPath, error)
		GetLabPathByID(ctx context.Context, id uuid.UUID) (*entity.LabPath, error)
		GetLabPathByName(ctx context.Context, name string) (*entity.LabPath, error)
		CreateLabPath(ctx context.Context, labPath *entity.LabPath) error
		UpdateLabPath(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*entity.LabPath, error)
		DeleteLabPath(ctx context.Context, id uuid.UUID) error

		// Prerequisite
		GetPrerequisiteByCourseAndRequire(ctx context.Context, courseID, requireID uuid.UUID) (*entity.Prerequisite, error)
		GetPrerequisiteByID(ctx context.Context, courseID, requireID uuid.UUID) (*entity.Prerequisite, error)
		CreatePrerequisite(ctx context.Context, prereq *entity.Prerequisite) error
		DeletePrerequisite(ctx context.Context, courseID, requireID uuid.UUID) error

		// Path Edge
		GetPathEdgeByID(ctx context.Context, id uuid.UUID) (*entity.PathEdge, error)
		GetPathEdgeByFromTo(ctx context.Context, fromID, toID uuid.UUID) (*entity.PathEdge, error)
		CreatePathEdge(ctx context.Context, edge *entity.PathEdge) error
		DeletePathEdge(ctx context.Context, id uuid.UUID) error

		// Lecture
		GetAllLectures(ctx context.Context) ([]entity.Lecture, error)
		GetLectureByID(ctx context.Context, id uuid.UUID) (*entity.Lecture, error)
		GetLectureByCode(ctx context.Context, code string) (*entity.Lecture, error)
		CreateLecture(ctx context.Context, lecture *entity.Lecture) error
		UpdateLecture(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*entity.Lecture, error)
	}

	adminRepository struct {
		db *gorm.DB
	}
)

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

// =========== COURSE ===========

func (r *adminRepository) GetAllCourses(ctx context.Context) ([]entity.Course, error) {
	var courses []entity.Course
	err := r.db.WithContext(ctx).Preload("LabPath").Find(&courses).Error
	return courses, err
}

func (r *adminRepository) GetCourseByID(ctx context.Context, id uuid.UUID) (*entity.Course, error) {
	var course entity.Course
	err := r.db.WithContext(ctx).Preload("LabPath").Where("id = ?", id).First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *adminRepository) GetCourseByCode(ctx context.Context, code string) (*entity.Course, error) {
	var course entity.Course
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&course).Error
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *adminRepository) CreateCourse(ctx context.Context, course *entity.Course) error {
	return r.db.WithContext(ctx).Create(course).Error
}

func (r *adminRepository) UpdateCourse(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*entity.Course, error) {
	if err := r.db.WithContext(ctx).Model(&entity.Course{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetCourseByID(ctx, id)
}

func (r *adminRepository) DeleteCourse(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Course{}).Error
}

// =========== LAB PATH ===========

func (r *adminRepository) GetAllLabPaths(ctx context.Context) ([]entity.LabPath, error) {
	var labPaths []entity.LabPath
	err := r.db.WithContext(ctx).Find(&labPaths).Error
	return labPaths, err
}

func (r *adminRepository) GetLabPathByID(ctx context.Context, id uuid.UUID) (*entity.LabPath, error) {
	var labPath entity.LabPath
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&labPath).Error
	if err != nil {
		return nil, err
	}
	return &labPath, nil
}

func (r *adminRepository) GetLabPathByName(ctx context.Context, name string) (*entity.LabPath, error) {
	var labPath entity.LabPath
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&labPath).Error
	if err != nil {
		return nil, err
	}
	return &labPath, nil
}

func (r *adminRepository) CreateLabPath(ctx context.Context, labPath *entity.LabPath) error {
	return r.db.WithContext(ctx).Create(labPath).Error
}

func (r *adminRepository) UpdateLabPath(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*entity.LabPath, error) {
	if err := r.db.WithContext(ctx).Model(&entity.LabPath{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetLabPathByID(ctx, id)
}

func (r *adminRepository) DeleteLabPath(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.LabPath{}).Error
}

// =========== PREREQUISITE ===========

func (r *adminRepository) GetPrerequisiteByCourseAndRequire(ctx context.Context, courseID, requireID uuid.UUID) (*entity.Prerequisite, error) {
	var prereq entity.Prerequisite
	err := r.db.WithContext(ctx).Where("course_id = ? AND require_id = ?", courseID, requireID).First(&prereq).Error
	if err != nil {
		return nil, err
	}
	return &prereq, nil
}

func (r *adminRepository) GetPrerequisiteByID(ctx context.Context, courseID, requireID uuid.UUID) (*entity.Prerequisite, error) {
	return r.GetPrerequisiteByCourseAndRequire(ctx, courseID, requireID)
}

func (r *adminRepository) CreatePrerequisite(ctx context.Context, prereq *entity.Prerequisite) error {
	return r.db.WithContext(ctx).Create(prereq).Error
}

func (r *adminRepository) DeletePrerequisite(ctx context.Context, courseID, requireID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("course_id = ? AND require_id = ?", courseID, requireID).Delete(&entity.Prerequisite{}).Error
}

// =========== PATH EDGE ===========

func (r *adminRepository) GetPathEdgeByID(ctx context.Context, id uuid.UUID) (*entity.PathEdge, error) {
	var edge entity.PathEdge
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&edge).Error
	if err != nil {
		return nil, err
	}
	return &edge, nil
}

func (r *adminRepository) GetPathEdgeByFromTo(ctx context.Context, fromID, toID uuid.UUID) (*entity.PathEdge, error) {
	var edge entity.PathEdge
	err := r.db.WithContext(ctx).Where("from_course_id = ? AND to_course_id = ?", fromID, toID).First(&edge).Error
	if err != nil {
		return nil, err
	}
	return &edge, nil
}

func (r *adminRepository) CreatePathEdge(ctx context.Context, edge *entity.PathEdge) error {
	return r.db.WithContext(ctx).Create(edge).Error
}

func (r *adminRepository) DeletePathEdge(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.PathEdge{}).Error
}

// =========== LECTURE ===========

func (r *adminRepository) GetAllLectures(ctx context.Context) ([]entity.Lecture, error) {
	var lectures []entity.Lecture
	err := r.db.WithContext(ctx).Find(&lectures).Error
	return lectures, err
}

func (r *adminRepository) GetLectureByID(ctx context.Context, id uuid.UUID) (*entity.Lecture, error) {
	var lecture entity.Lecture
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&lecture).Error
	if err != nil {
		return nil, err
	}
	return &lecture, nil
}

func (r *adminRepository) GetLectureByCode(ctx context.Context, code string) (*entity.Lecture, error) {
	var lecture entity.Lecture
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&lecture).Error
	if err != nil {
		return nil, err
	}
	return &lecture, nil
}

func (r *adminRepository) CreateLecture(ctx context.Context, lecture *entity.Lecture) error {
	return r.db.WithContext(ctx).Create(lecture).Error
}

func (r *adminRepository) UpdateLecture(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*entity.Lecture, error) {
	if err := r.db.WithContext(ctx).Model(&entity.Lecture{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetLectureByID(ctx, id)
}
