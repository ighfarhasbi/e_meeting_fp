package models

type Dashboard struct {
	TotalRoom        int              `json:"totalRoom"`
	TotalVisitor     int              `json:"totalVisitor"`
	TotalReservation int              `json:"totalReservation"`
	TotalOmzet       float64          `json:"totalOmzet"`
	Rooms            []RoomsDashboard `json:"rooms"`
}

type RoomsDashboard struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Omzet             float64 `json:"omzet"`
	PercentegeOfUsage int     `json:"percentageOfUsage"`
}
