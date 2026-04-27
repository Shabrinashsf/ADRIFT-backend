package service

import "ADRIFT-backend/internal/api/repository"

type (
	FRSService interface {
	}

	frsService struct {
		frsRepo repository.FRSRepository
	}
)

func NewFRSService(frsRepo repository.FRSRepository) FRSService {
	return &frsService{
		frsRepo: frsRepo,
	}
}
