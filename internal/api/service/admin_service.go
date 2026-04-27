package service

import "ADRIFT-backend/internal/api/repository"

type (
	AdminService interface {
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
