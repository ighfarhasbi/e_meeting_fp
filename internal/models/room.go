package models

import "github.com/shopspring/decimal"

type Room struct {
	ID           int             `json:"id" db:"rooms_id"`
	Name         string          `json:"name" db:"rooms_name"`
	Type         string          `json:"type" db:"rooms_type"`
	PricePerHour decimal.Decimal `json:"pricePerHour" db:"rooms_price_perhour"`
	Capacity     int             `json:"capacity" db:"rooms_capacity"`
	ImgPath      string          `json:"imgUrl" db:"rooms_img_path"`
	CreatedAt    string          `json:"createdAt" db:"created_at"`
	UpdatedAt    string          `json:"updatedAt" db:"updated_at"`
}

// CU -> Create or Update
type CURoomRequest struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	PricePerHour float64 `json:"pricePerHour"`
	Capacity     int     `json:"capacity"`
	ImgPath      string  `json:"imgUrl"`
}

// bedanya di price per hour harus float
type RoomResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	PricePerHour float64 `json:"pricePerHour"`
	Capacity     int     `json:"capacity"`
	ImgPath      string  `json:"imgUrl"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type RoomSchedule struct {
	// Name      string `json:"name"`
	Status    string `json:"status"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}
type RoomScheduleResponse struct {
	RoomName    string         `json:"roomName"`
	Schedule    []RoomSchedule `json:"schedule"`
	TotalBooked int            `json:"totalBooked"`
}

type RoomScheduleAdmin struct {
	// Name      string `json:"name"`
	Company        string `json:"company"`
	StartTime      string `json:"startTime"`
	EndTime        string `json:"endTime"`
	Status         string `json:"status"`
	StatusProgress string `json:"statusProgress"`
}
type RoomScheduleAdminResponse struct {
	RoomName string `json:"roomName"`
	// Company  string              `json:"company"`
	Schedule []RoomScheduleAdmin `json:"schedule"`
}
