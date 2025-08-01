package models

import "github.com/shopspring/decimal"

type Snacks struct {
	ID       int             `json:"id" db:"snacks_id"`
	Name     string          `json:"name" db:"name"`
	Price    decimal.Decimal `json:"price" db:"price"`
	Category string          `json:"category" db:"category"`
}
