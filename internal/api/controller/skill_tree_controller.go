package controller

import (
	"ADRIFT-backend/internal/api/service"
	"ADRIFT-backend/internal/pkg/validate"
)

type (
	SkillTreeController interface {
	}

	skillTreeController struct {
		skillTreeService service.SkillTreeService
		validator        *validate.Validator
	}
)

func NewSkillTreeController(sts service.SkillTreeService, validator *validate.Validator) SkillTreeController {
	return &skillTreeController{
		skillTreeService: sts,
		validator:        validator,
	}
}
