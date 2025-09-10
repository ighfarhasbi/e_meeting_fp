package repository

import (
	"e_meeting/internal/entity"
	"e_meeting/internal/models/response"
)

type RoomsRepository interface {
	GetAllRooms(roomName string, roomType string, capacity int, pageSize int, offset int) ([]entity.Rooms, error)
	CountTotalRooms(roomName string, roomType string, capacity int) (int, error)
	GetRoomsByIDAndDate(id int, date string) (*response.RoomsSchedules, error)
	CreateRoom(room *entity.Rooms) error
	UpdateRoomByID(room *entity.Rooms) error
	DeleteRoomByID(id int) error
}
