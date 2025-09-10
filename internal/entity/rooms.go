package entity

import "github.com/shopspring/decimal"

type Rooms struct {
	ID           int             `json:"id" db:"rooms_id"`
	Name         string          `json:"name" db:"name"`
	Type         string          `json:"type" db:"type"`
	PricePerHour decimal.Decimal `json:"pricePerHour" db:"price_perhour"`
	Capacity     int             `json:"capacity" db:"capacity"`
	ImgUrl       string          `json:"imgUrl" db:"img_path"`
	CreatedAt    string          `json:"createdAt" db:"created_at"`
	UpdatedAt    string          `json:"updatedAt" db:"updated_at"`
}
