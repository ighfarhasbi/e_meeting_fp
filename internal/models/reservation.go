package models

type PersonalDataCalculation struct {
	Name         string `json:"name"`
	NoHp         string `json:"noHp"`
	Company      string `json:"company"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	Duration     int    `json:"durationHours"`
	Participants int    `json:"participants"`
}

type SnacksCalculation struct {
	ID       int     `json:"id"`
	Category string  `json:"category"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
}

type RoomCalculation struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	PricePerHour  float64           `json:"pricePerHour"`
	Capacity      int               `json:"capacity"`
	ImgPath       string            `json:"imgPath"`
	SubTotalSnack float64           `json:"subTotalSnack"`
	SubTotalRoom  float64           `json:"subTotalRoom"`
	Snacks        SnacksCalculation `json:"snacks"`
}

type CalculationResponse struct {
	Rooms        []RoomCalculation       `json:"rooms"`
	PersonalData PersonalDataCalculation `json:"personalData"`
	Total        float64                 `json:"total"`
}

type ReservationRequest struct {
	UserID      int               `json:"userID"`
	Name        string            `json:"name"`
	PhoneNumber string            `json:"phoneNumber"`
	Company     string            `json:"company"`
	Notes       string            `json:"notes"`
	Rooms       []RoomReservation `json:"rooms"`
}
type RoomReservation struct {
	ID           int    `json:"id"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	Participants int    `json:"participants"`
	SnackID      int    `json:"snackID"`
}
