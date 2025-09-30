package request

type CreateRoomRequest struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	PricePerHour float64 `json:"pricePerHour"`
	Capacity     int     `json:"capacity"`
	ImgUrl       string  `json:"imgUrl"`
}
type UpdateRoomRequest struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	PricePerHour float64 `json:"pricePerHour"`
	Capacity     int     `json:"capacity"`
	ImgUrl       string  `json:"imgUrl"`
}
