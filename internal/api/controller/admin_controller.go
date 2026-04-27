package controller

import (
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/pkg/validate"
)

type (
	AdminController interface {
	}

	adminController struct {
		adminService service.AdminService
		validator    *validate.Validator
	}
)

func NewAdminController(ads service.AdminService, validator *validate.Validator) AdminController {
	return &adminController{
		adminService: ads,
		validator:    validator,
	}
}
