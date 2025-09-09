package response

type Schedules struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status    string `json:"status"`
}

type RoomsSchedules struct {
	RoomName    string      `json:"roomName"`
	Schedules   []Schedules `json:"schedules"`
	TotalBooked int         `json:"totalBooked"`
}
