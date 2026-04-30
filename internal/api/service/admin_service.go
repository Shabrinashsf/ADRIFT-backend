package service

import (
	"context"

	"ADRIFT-backend/internal/api/repository"
	"ADRIFT-backend/internal/dto"
	"ADRIFT-backend/internal/entity"

	"github.com/google/uuid"
)

type (
	AdminService interface {
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
	return &adminService{adminRepo: adminRepo}
}

// =========== LAB PATH ===========

func (s *adminService) GetAllLabPaths(ctx context.Context) ([]dto.AdminLabPathResponse, error) {
	labPaths, err := s.adminRepo.GetAllLabPaths(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AdminLabPathResponse, 0, len(labPaths))
	for _, lp := range labPaths {
		result = append(result, dto.AdminLabPathResponse{
			ID:    lp.ID.String(),
			Name:  lp.Name,
			Color: lp.Color,
		})
	}
	return result, nil
}

func (s *adminService) CreateLabPath(ctx context.Context, req dto.CreateLabPathRequest) (dto.AdminLabPathResponse, error) {
	_, err := s.adminRepo.GetLabPathByName(ctx, req.Name)
	if err == nil {
		return dto.AdminLabPathResponse{}, dto.ErrLabPathNameExists
	}

	id, _ := uuid.NewV7()
	labPath := &entity.LabPath{
		ID:    id,
		Name:  req.Name,
		Color: req.Color,
	}

	if err := s.adminRepo.CreateLabPath(ctx, labPath); err != nil {
		return dto.AdminLabPathResponse{}, err
	}

	return dto.AdminLabPathResponse{
		ID:    labPath.ID.String(),
		Name:  labPath.Name,
		Color: labPath.Color,
	}, nil
}

func (s *adminService) UpdateLabPath(ctx context.Context, id uuid.UUID, req dto.UpdateLabPathRequest) (dto.AdminLabPathResponse, error) {
	_, err := s.adminRepo.GetLabPathByID(ctx, id)
	if err != nil {
		return dto.AdminLabPathResponse{}, dto.ErrLabPathNotFound
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}

	updated, err := s.adminRepo.UpdateLabPath(ctx, id, updates)
	if err != nil {
		return dto.AdminLabPathResponse{}, err
	}

	return dto.AdminLabPathResponse{
		ID:    updated.ID.String(),
		Name:  updated.Name,
		Color: updated.Color,
	}, nil
}

func (s *adminService) DeleteLabPath(ctx context.Context, id uuid.UUID) error {
	_, err := s.adminRepo.GetLabPathByID(ctx, id)
	if err != nil {
		return dto.ErrLabPathNotFound
	}
	return s.adminRepo.DeleteLabPath(ctx, id)
}

// =========== PREREQUISITE ===========

func (s *adminService) CreatePrerequisite(ctx context.Context, req dto.CreatePrerequisiteRequest) (dto.AdminPrerequisiteResponse, error) {
	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		return dto.AdminPrerequisiteResponse{}, dto.ErrAdminCourseNotFound
	}
	requireID, err := uuid.Parse(req.RequireID)
	if err != nil {
		return dto.AdminPrerequisiteResponse{}, dto.ErrAdminCourseNotFound
	}

	// Check duplicate
	_, err = s.adminRepo.GetPrerequisiteByCourseAndRequire(ctx, courseID, requireID)
	if err == nil {
		return dto.AdminPrerequisiteResponse{}, dto.ErrPrerequisiteExists
	}

	id, _ := uuid.NewV7()
	prereq := &entity.Prerequisite{
		ID:        id,
		CourseID:  courseID,
		RequireID: requireID,
	}

	if err := s.adminRepo.CreatePrerequisite(ctx, prereq); err != nil {
		return dto.AdminPrerequisiteResponse{}, err
	}

	return dto.AdminPrerequisiteResponse{
		ID:        prereq.ID.String(),
		CourseID:  prereq.CourseID.String(),
		RequireID: prereq.RequireID.String(),
	}, nil
}

func (s *adminService) DeletePrerequisite(ctx context.Context, courseID, requireID uuid.UUID) error {
	_, err := s.adminRepo.GetPrerequisiteByCourseAndRequire(ctx, courseID, requireID)
	if err != nil {
		return dto.ErrPrerequisiteNotFound
	}
	return s.adminRepo.DeletePrerequisite(ctx, courseID, requireID)
}

// =========== PATH EDGE ===========

func (s *adminService) CreatePathEdge(ctx context.Context, req dto.CreatePathEdgeRequest) (dto.AdminPathEdgeResponse, error) {
	fromID, err := uuid.Parse(req.FromCourseID)
	if err != nil {
		return dto.AdminPathEdgeResponse{}, dto.ErrAdminCourseNotFound
	}
	toID, err := uuid.Parse(req.ToCourseID)
	if err != nil {
		return dto.AdminPathEdgeResponse{}, dto.ErrAdminCourseNotFound
	}

	// Check duplicate
	_, err = s.adminRepo.GetPathEdgeByFromTo(ctx, fromID, toID)
	if err == nil {
		return dto.AdminPathEdgeResponse{}, dto.ErrPathEdgeExists
	}

	id, _ := uuid.NewV7()
	edge := &entity.PathEdge{
		ID:           id,
		FromCourseID: fromID,
		ToCourseID:   toID,
	}

	if err := s.adminRepo.CreatePathEdge(ctx, edge); err != nil {
		return dto.AdminPathEdgeResponse{}, err
	}

	return dto.AdminPathEdgeResponse{
		ID:           edge.ID.String(),
		FromCourseID: edge.FromCourseID.String(),
		ToCourseID:   edge.ToCourseID.String(),
	}, nil
}

func (s *adminService) DeletePathEdge(ctx context.Context, id uuid.UUID) error {
	_, err := s.adminRepo.GetPathEdgeByID(ctx, id)
	if err != nil {
		return dto.ErrPathEdgeNotFound
	}
	return s.adminRepo.DeletePathEdge(ctx, id)
}

// =========== LECTURE ===========

func (s *adminService) GetAllLectures(ctx context.Context) ([]dto.AdminLectureResponse, error) {
	lectures, err := s.adminRepo.GetAllLectures(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.AdminLectureResponse, 0, len(lectures))
	for _, l := range lectures {
		result = append(result, dto.AdminLectureResponse{
			ID:   l.ID.String(),
			Code: l.Code,
			Name: l.Name,
		})
	}
	return result, nil
}

func (s *adminService) CreateLecture(ctx context.Context, req dto.CreateLectureRequest) (dto.AdminLectureResponse, error) {
	id, _ := uuid.NewV7()
	lecture := &entity.Lecture{
		ID:   id,
		Code: req.Code,
		Name: req.Name,
	}

	if err := s.adminRepo.CreateLecture(ctx, lecture); err != nil {
		return dto.AdminLectureResponse{}, err
	}

	return dto.AdminLectureResponse{
		ID:   lecture.ID.String(),
		Code: lecture.Code,
		Name: lecture.Name,
	}, nil
}

func (s *adminService) UpdateLecture(ctx context.Context, id uuid.UUID, req dto.UpdateLectureRequest) (dto.AdminLectureResponse, error) {
	_, err := s.adminRepo.GetLectureByID(ctx, id)
	if err != nil {
		return dto.AdminLectureResponse{}, dto.ErrLectureNotFound
	}

	updates := map[string]interface{}{}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}

	updated, err := s.adminRepo.UpdateLecture(ctx, id, updates)
	if err != nil {
		return dto.AdminLectureResponse{}, err
	}

	return dto.AdminLectureResponse{
		ID:   updated.ID.String(),
		Code: updated.Code,
		Name: updated.Name,
	}, nil
}
