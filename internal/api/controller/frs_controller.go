package controller

import (
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/pkg/validate"
)

type (
	FRSController interface {
	}

	frsController struct {
		frsService service.FRSService
		validator  *validate.Validator
	}
)

func NewFRSController(frsService service.FRSService, validator *validate.Validator) FRSController {
	return &frsController{
		frsService: frsService,
		validator:  validator,
	}
}
