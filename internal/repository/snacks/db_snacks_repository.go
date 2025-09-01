package repository

import (
	"database/sql"
	"e_meeting/internal/entity"
)

type DBSnacksRepository struct {
	DB *sql.DB
}

func NewDBSnacksRepository(db *sql.DB) *DBSnacksRepository {
	return &DBSnacksRepository{
		DB: db,
	}
}

func (r *DBSnacksRepository) GetAllSnacks() ([]entity.Snacks, error) {
	rows, err := r.DB.Query("SELECT snacks_id, name, price, category FROM snacks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snacksList []entity.Snacks
	for rows.Next() {
		var snack entity.Snacks
		if err := rows.Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Category); err != nil {
			if err == sql.ErrNoRows {
				return []entity.Snacks{}, nil
			}
			return nil, err
		}
		snacksList = append(snacksList, snack)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return snacksList, nil
}
