package usecase

import (
	"e_meeting/internal/entity"
	repository "e_meeting/internal/repository/snacks"
)

type SnacksUsecase struct {
	repo repository.SnacksRepository
}

func NewSnacksUsecase(r repository.SnacksRepository) *SnacksUsecase {
	return &SnacksUsecase{repo: r}
}

func (uc *SnacksUsecase) SnacksList() ([]entity.Snacks, error) {
	snackList, err := uc.repo.GetAllSnacks()
	if err != nil {
		return nil, err
	}
	return snackList, nil
}
