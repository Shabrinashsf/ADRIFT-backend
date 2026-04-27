package service

import "ADRIFT-backend/internal/api/repository"

type (
	SkillTreeService interface {
	}

	skillTreeService struct {
		skillTreeRepo repository.SkillTreeRepository
	}
)

func NewSkillTreeService(stRepo repository.SkillTreeRepository) SkillTreeService {
	return &skillTreeService{
		skillTreeRepo: stRepo,
	}
}
