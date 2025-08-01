package models

import "github.com/shopspring/decimal"

type Room struct {
	ID           int             `json:"id" db:"rooms_id"`
	Name         string          `json:"name" db:"rooms_name"`
	Type         string          `json:"type" db:"rooms_type"`
	PricePerHour decimal.Decimal `json:"price_per_hour" db:"rooms_price_perhour"`
	Capacity     int             `json:"capacity" db:"rooms_capacity"`
	ImgUrl       string          `json:"img_url" db:"rooms_img_path"`
	CreatedAt    string          `json:"created_at" db:"created_at"`
	UpdatedAt    string          `json:"updated_at" db:"updated_at"`
}
