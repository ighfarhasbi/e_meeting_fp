package repository

import "e_meeting/internal/entity"

type SnacksRepository interface {
	GetAllSnacks() ([]entity.Snacks, error)
}
