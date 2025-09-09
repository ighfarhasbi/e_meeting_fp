package repository

import (
	"e_meeting/internal/entity"
	"e_meeting/internal/models/response"
)

type RoomsRepository interface {
	GetAllRooms() ([]entity.Rooms, error)
	GetRoomsByIDAndDate(id int, date string) (*response.RoomsSchedules, error)
	CreateRoom(room *entity.Rooms) error
	UpdateRoomByID(room *entity.Rooms) error
	DeleteRoomByID(id int) error
}
