package service

import (
	"context"
	"fmt"
	"time"

	"ADRIFT-backend/internal/api/repository"
	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	AdminService interface {
		// Course
		ListCourseGroups(ctx context.Context) ([]dto.AdminCourseGroupResponse, error)
		ListCoursesBySemester(ctx context.Context, semester int, courseName string) ([]dto.AdminCourseResponse, error)
		CreateCourse(ctx context.Context, req dto.AdminCreateCourseRequest) error
		UpdateCourse(ctx context.Context, courseID string, req dto.AdminUpdateCourseRequest) (dto.AdminUpdateCourseResponse, error)
		DeleteCourse(ctx context.Context, courseID string) error

		// Schedule
		ListScheduleGroups(ctx context.Context) ([]dto.AdminScheduleGroupResponse, error)
		ListSchedulesByFilter(ctx context.Context, academicYear, term, prodi, semester, courseName string) ([]dto.AdminScheduleResponse, error)
		CreateSchedule(ctx context.Context, req dto.AdminCreateScheduleRequest) error
		UpdateSchedule(ctx context.Context, scheduleID string, req dto.AdminUpdateScheduleRequest) (dto.AdminUpdateScheduleResponse, error)
		DeleteSchedule(ctx context.Context, scheduleID string) error
		// Lab Path
		GetAllLabPaths(ctx context.Context) ([]dto.AdminLabPathResponse, error)
		CreateLabPath(ctx context.Context, req dto.CreateLabPathRequest) (dto.AdminLabPathResponse, error)
		UpdateLabPath(ctx context.Context, id uuid.UUID, req dto.UpdateLabPathRequest) (dto.AdminLabPathResponse, error)
		DeleteLabPath(ctx context.Context, id uuid.UUID) error

		// Prerequisite
		CreatePrerequisite(ctx context.Context, req dto.CreatePrerequisiteRequest) (dto.AdminPrerequisiteResponse, error)
		DeletePrerequisite(ctx context.Context, courseID, requireID uuid.UUID) error

		// Path Edge
		CreatePathEdge(ctx context.Context, req dto.CreatePathEdgeRequest) (dto.AdminPathEdgeResponse, error)
		DeletePathEdge(ctx context.Context, id uuid.UUID) error

		// Lecture
		GetAllLectures(ctx context.Context) ([]dto.AdminLectureResponse, error)
		CreateLecture(ctx context.Context, req dto.CreateLectureRequest) (dto.AdminLectureResponse, error)
		UpdateLecture(ctx context.Context, id uuid.UUID, req dto.UpdateLectureRequest) (dto.AdminLectureResponse, error)
	}

	adminService struct {
		adminRepo repository.AdminRepository
	}
)

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{
		adminRepo: adminRepo,
	}
}

func formatTimeToWIB(t time.Time) string {
	return t.Add(7 * time.Hour).Format("15:04")
}

func parseTimeFromWIB(tStr string) (time.Time, error) {
	parsed, err := time.Parse("15:04", tStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("format waktu harus HH:MM")
	}
	utcTime := parsed.Add(-7 * time.Hour)
	baseTime := time.Date(2000, 1, 1, utcTime.Hour(), utcTime.Minute(), 0, 0, time.UTC)
	return baseTime, nil
}

func (s *adminService) ListCourseGroups(ctx context.Context) ([]dto.AdminCourseGroupResponse, error) {
	groups, err := s.adminRepo.GetCoursesGroupedBySemester(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminCourseGroupResponse, len(groups))
	for i, g := range groups {
		result[i] = dto.AdminCourseGroupResponse{
			Semester:    g.Semester,
			TotalCourse: g.TotalCourse,
		}
	}
	return result, nil
}

func (s *adminService) ListCoursesBySemester(ctx context.Context, semester int, courseName string) ([]dto.AdminCourseResponse, error) {
	courses, err := s.adminRepo.GetCoursesBySemester(ctx, semester, courseName)
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminCourseResponse, len(courses))
	for i, c := range courses {
		result[i] = dto.AdminCourseResponse{
			ID:          c.ID.String(),
			Name:        c.Name,
			Code:        c.Code,
			Credit:      c.Credit,
			Semester:    c.Semester,
			IsElective:  c.IsElective,
			Description: c.Description,
			Lab:         c.Lab,
		}
	}
	return result, nil
}

func (s *adminService) CreateCourse(ctx context.Context, req dto.AdminCreateCourseRequest) error {
	_, err := s.adminRepo.GetCourseByCode(ctx, req.Code)
	if err == nil {
		return dto.ErrCourseDuplicate
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	course := entity.Course{
		ID:          uuid.New(),
		Code:        req.Code,
		Name:        req.Name,
		Credit:      req.Credit,
		Semester:    req.Semester,
		IsElective:  req.IsElective,
		Description: req.Description,
		Lab:         req.Lab,
	}

	_, err = s.adminRepo.CreateCourse(ctx, course)
	return err
}

func (s *adminService) UpdateCourse(ctx context.Context, courseID string, req dto.AdminUpdateCourseRequest) (dto.AdminUpdateCourseResponse, error) {
	id, err := uuid.Parse(courseID)
	if err != nil {
		return dto.AdminUpdateCourseResponse{}, dto.ErrInvalidUUID
	}

	existing, err := s.adminRepo.GetCourseByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.AdminUpdateCourseResponse{}, dto.ErrAdminCourseNotFound
		}
		return dto.AdminUpdateCourseResponse{}, err
	}

	if req.Code != nil && *req.Code != existing.Code {
		dup, dupErr := s.adminRepo.GetCourseByCode(ctx, *req.Code)
		if dupErr == nil && dup.ID != id {
			return dto.AdminUpdateCourseResponse{}, dto.ErrCourseDuplicate
		}
		if dupErr != nil && dupErr != gorm.ErrRecordNotFound {
			return dto.AdminUpdateCourseResponse{}, dupErr
		}
	}

	updates := make(map[string]interface{})
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Credit != nil {
		updates["credit"] = *req.Credit
	}
	if req.Semester != nil {
		updates["semester"] = *req.Semester
	}
	if req.IsElective != nil {
		updates["is_elective"] = *req.IsElective
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Lab != nil {
		updates["lab"] = *req.Lab
	}

	updated, err := s.adminRepo.UpdateCourse(ctx, id, updates)
	if err != nil {
		return dto.AdminUpdateCourseResponse{}, err
	}

	return dto.AdminUpdateCourseResponse{
		ID:          updated.ID.String(),
		Code:        updated.Code,
		Name:        updated.Name,
		Credit:      updated.Credit,
		Semester:    updated.Semester,
		IsElective:  updated.IsElective,
		Description: updated.Description,
		Lab:         updated.Lab,
	}, nil
}

func (s *adminService) DeleteCourse(ctx context.Context, courseID string) error {
	id, err := uuid.Parse(courseID)
	if err != nil {
		return dto.ErrInvalidUUID
	}

	err = s.adminRepo.SoftDeleteCourse(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.ErrAdminCourseNotFound
		}
		return err
	}
	return nil
}

func (s *adminService) ListScheduleGroups(ctx context.Context) ([]dto.AdminScheduleGroupResponse, error) {
	groups, err := s.adminRepo.GetScheduleGroups(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminScheduleGroupResponse, len(groups))
	for i, g := range groups {
		result[i] = dto.AdminScheduleGroupResponse{
			AcademicYear: g.AcademicYear,
			Term:         g.Term,
			Prodi:        g.Prodi,
			Semester:     g.Semester,
		}
	}
	return result, nil
}

func (s *adminService) ListSchedulesByFilter(ctx context.Context, academicYear, term, prodi, semester, courseName string) ([]dto.AdminScheduleResponse, error) {
	semesterInt := 0
	_, err := fmt.Sscanf(semester, "%d", &semesterInt)
	if err != nil {
		return nil, dto.ErrInvalidSemester
	}

	schedules, err := s.adminRepo.GetSchedulesByFilter(ctx, academicYear, term, entity.ProdiType(prodi), semesterInt, courseName)
	if err != nil {
		return nil, err
	}

	result := make([]dto.AdminScheduleResponse, len(schedules))
	for i, sch := range schedules {
		lectureName := ""
		if sch.Lecture != nil {
			lectureName = sch.Lecture.Name
		}

		result[i] = dto.AdminScheduleResponse{
			ID:           sch.ID.String(),
			CourseName:   sch.CourseName,
			LectureName:  lectureName,
			Class:        sch.Class,
			Day:          string(sch.Day),
			StartTime:    formatTimeToWIB(sch.StartTime),
			EndTime:      formatTimeToWIB(sch.EndTime),
			Room:         sch.Room,
			Semester:     sch.Semester,
			AcademicYear: sch.AcademicYear,
			Capacity:     sch.Capacity,
			SKS:          sch.SKS,
			Prodi:        string(sch.Prodi),
			Term:         string(sch.Term),
		}
	}
	return result, nil
}

func (s *adminService) CreateSchedule(ctx context.Context, req dto.AdminCreateScheduleRequest) error {
	startTime, err := parseTimeFromWIB(req.StartTime)
	if err != nil {
		return dto.ErrInvalidTimeFormat
	}

	endTime, err := parseTimeFromWIB(req.EndTime)
	if err != nil {
		return dto.ErrInvalidTimeFormat
	}

	lectureID, err := uuid.Parse(req.LectureID)
	if err != nil {
		return dto.ErrInvalidUUID
	}

	duplicate, err := s.adminRepo.CheckScheduleDuplicate(ctx, req.CourseName, req.Class, req.AcademicYear, req.Term, entity.ProdiType(req.Prodi), req.Semester)
	if err != nil {
		return err
	}
	if duplicate {
		return dto.ErrScheduleDuplicate
	}

	schedule := entity.Schedule{
		ID:           uuid.New(),
		CourseName:   req.CourseName,
		LectureID:    lectureID,
		Class:        req.Class,
		Day:          entity.Day(req.Day),
		StartTime:    startTime,
		EndTime:      endTime,
		Room:         req.Room,
		Semester:     req.Semester,
		AcademicYear: req.AcademicYear,
		Capacity:     req.Capacity,
		SKS:          req.SKS,
		Prodi:        entity.ProdiType(req.Prodi),
		Term:         entity.TermSemester(req.Term),
	}

	_, err = s.adminRepo.CreateSchedule(ctx, schedule)
	return err
}

func (s *adminService) UpdateSchedule(ctx context.Context, scheduleID string, req dto.AdminUpdateScheduleRequest) (dto.AdminUpdateScheduleResponse, error) {
	id, err := uuid.Parse(scheduleID)
	if err != nil {
		return dto.AdminUpdateScheduleResponse{}, dto.ErrInvalidUUID
	}

	existing, err := s.adminRepo.GetScheduleByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.AdminUpdateScheduleResponse{}, dto.ErrScheduleNotFound
		}
		return dto.AdminUpdateScheduleResponse{}, err
	}

	courseName := existing.CourseName
	class := existing.Class
	academicYear := existing.AcademicYear
	term := string(existing.Term)
	prodi := string(existing.Prodi)
	semester := existing.Semester

	if req.CourseName != nil {
		courseName = *req.CourseName
	}
	if req.Class != nil {
		class = *req.Class
	}
	if req.AcademicYear != nil {
		academicYear = *req.AcademicYear
	}
	if req.Term != nil {
		term = *req.Term
	}
	if req.Prodi != nil {
		prodi = *req.Prodi
	}
	if req.Semester != nil {
		semester = *req.Semester
	}

	duplicate, err := s.adminRepo.CheckScheduleDuplicateExcludeID(ctx, id, courseName, class, academicYear, term, entity.ProdiType(prodi), semester)
	if err != nil {
		return dto.AdminUpdateScheduleResponse{}, err
	}
	if duplicate {
		return dto.AdminUpdateScheduleResponse{}, dto.ErrScheduleDuplicate
	}

	updates := make(map[string]interface{})
	if req.CourseName != nil {
		updates["course_name"] = *req.CourseName
	}
	if req.LectureID != nil {
		lid, parseErr := uuid.Parse(*req.LectureID)
		if parseErr != nil {
			return dto.AdminUpdateScheduleResponse{}, dto.ErrInvalidUUID
		}
		updates["lecture_id"] = lid
	}
	if req.Class != nil {
		updates["class"] = *req.Class
	}
	if req.Day != nil {
		updates["day"] = entity.Day(*req.Day)
	}
	if req.StartTime != nil {
		st, parseErr := parseTimeFromWIB(*req.StartTime)
		if parseErr != nil {
			return dto.AdminUpdateScheduleResponse{}, dto.ErrInvalidTimeFormat
		}
		updates["start_time"] = st
	}
	if req.EndTime != nil {
		et, parseErr := parseTimeFromWIB(*req.EndTime)
		if parseErr != nil {
			return dto.AdminUpdateScheduleResponse{}, dto.ErrInvalidTimeFormat
		}
		updates["end_time"] = et
	}
	if req.Room != nil {
		updates["room"] = *req.Room
	}
	if req.Semester != nil {
		updates["semester"] = *req.Semester
	}
	if req.AcademicYear != nil {
		updates["academic_year"] = *req.AcademicYear
	}
	if req.Capacity != nil {
		updates["capacity"] = *req.Capacity
	}
	if req.SKS != nil {
		updates["sks"] = *req.SKS
	}
	if req.Prodi != nil {
		updates["prodi"] = entity.ProdiType(*req.Prodi)
	}
	if req.Term != nil {
		updates["term"] = entity.TermSemester(*req.Term)
	}

	updated, err := s.adminRepo.UpdateSchedule(ctx, id, updates)
	if err != nil {
		return dto.AdminUpdateScheduleResponse{}, err
	}

	return dto.AdminUpdateScheduleResponse{
		ID:           updated.ID.String(),
		CourseName:   updated.CourseName,
		LectureID:    updated.LectureID.String(),
		Class:        updated.Class,
		Day:          string(updated.Day),
		StartTime:    formatTimeToWIB(updated.StartTime),
		EndTime:      formatTimeToWIB(updated.EndTime),
		Room:         updated.Room,
		Semester:     updated.Semester,
		AcademicYear: updated.AcademicYear,
		Capacity:     updated.Capacity,
		SKS:          updated.SKS,
		Prodi:        string(updated.Prodi),
		Term:         string(updated.Term),
	}, nil
}

func (s *adminService) DeleteSchedule(ctx context.Context, scheduleID string) error {
	id, err := uuid.Parse(scheduleID)
	if err != nil {
		return dto.ErrInvalidUUID
	}

	err = s.adminRepo.DeleteSchedule(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.ErrScheduleNotFound
		}
		return err
	}
	return nil
}
