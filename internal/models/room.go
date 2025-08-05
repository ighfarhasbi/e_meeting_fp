package models

import "github.com/shopspring/decimal"

type Room struct {
	ID           int             `json:"id" db:"rooms_id"`
	Name         string          `json:"name" db:"rooms_name"`
	Type         string          `json:"type" db:"rooms_type"`
	PricePerHour decimal.Decimal `json:"price_per_hour" db:"rooms_price_perhour"`
	Capacity     int             `json:"capacity" db:"rooms_capacity"`
	ImgPath      string          `json:"img_path" db:"rooms_img_path"`
	CreatedAt    string          `json:"created_at" db:"created_at"`
	UpdatedAt    string          `json:"updated_at" db:"updated_at"`
}

// CU -> Create or Update
type CURoomRequest struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	PricePerHour float64 `json:"price_per_hour"`
	Capacity     int     `json:"capacity"`
	ImgPath      string  `json:"img_path"`
}

// bedanya di price per hour harus float
type RoomResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	PricePerHour float64 `json:"price_per_hour"`
	Capacity     int     `json:"capacity"`
	ImgPath      string  `json:"img_path"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}
