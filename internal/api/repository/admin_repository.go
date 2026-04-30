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
)

type (
	AdminRepository interface {
		// Course
		GetCoursesGroupedBySemester(ctx context.Context) ([]CourseGroupResult, error)
		GetCoursesBySemester(ctx context.Context, semester int, courseName string) ([]entity.Course, error)
		GetCourseByCode(ctx context.Context, code string) (entity.Course, error)
		GetCourseByID(ctx context.Context, id uuid.UUID) (entity.Course, error)
		CreateCourse(ctx context.Context, course entity.Course) (entity.Course, error)
		UpdateCourse(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (entity.Course, error)
		SoftDeleteCourse(ctx context.Context, id uuid.UUID) error
		
		//Schedule
		GetScheduleGroups(ctx context.Context) ([]ScheduleGroupResult, error)
		GetSchedulesByFilter(ctx context.Context, academicYear, term string, prodi entity.ProdiType, semester int, courseName string) ([]entity.Schedule, error)
		GetScheduleByID(ctx context.Context, id uuid.UUID) (entity.Schedule, error)
		CheckScheduleDuplicate(ctx context.Context, courseName, class, academicYear, term string, prodi entity.ProdiType, semester int) (bool, error)
		CheckScheduleDuplicateExcludeID(ctx context.Context, excludeID uuid.UUID, courseName, class, academicYear, term string, prodi entity.ProdiType, semester int) (bool, error)
		CreateSchedule(ctx context.Context, schedule entity.Schedule) (entity.Schedule, error)
		UpdateSchedule(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (entity.Schedule, error)
		DeleteSchedule(ctx context.Context, id uuid.UUID) error
		
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