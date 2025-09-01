package usecase

import (
	"e_meeting/internal/entity"
	repository "e_meeting/internal/repository/snacks"
)

type SnacksUsecase struct {
	repo *repository.DBSnacksRepository
}

func NewSnacksUsecase(r *repository.DBSnacksRepository) *SnacksUsecase {
	return &SnacksUsecase{repo: r}
}

func (uc *SnacksUsecase) SnacksList() ([]entity.Snacks, error) {
	snackList, err := uc.repo.GetAllSnacks()
	if err != nil {
		return nil, err
	}
	return snackList, nil
}
