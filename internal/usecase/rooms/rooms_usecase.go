package usecase

import (
	"e_meeting/internal/entity"
	"e_meeting/internal/models/response"
	repository "e_meeting/internal/repository/rooms"
)

type RoomsUsecase struct {
	repo repository.RoomsRepository
}

func NewRoomsUsecase(r repository.RoomsRepository) *RoomsUsecase {
	return &RoomsUsecase{repo: r}
}

func (uc *RoomsUsecase) RoomsList() ([]entity.Rooms, error) {
	roomsList, err := uc.repo.GetAllRooms()
	if err != nil {
		return nil, err
	}
	return roomsList, nil
}

func (uc *RoomsUsecase) GetRoomSchedule(id int, date string) (*response.RoomsSchedules, error) {
	room, err := uc.repo.GetRoomsByIDAndDate(id, date)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (uc *RoomsUsecase) CreateRoom(room *entity.Rooms) error {
	if err := uc.repo.CreateRoom(room); err != nil {
		return err
	}
	return nil
}

func (uc *RoomsUsecase) UpdateRoom(room *entity.Rooms) error {

	if err := uc.repo.UpdateRoomByID(room); err != nil {
		return err
	}
	return nil
}

func (uc *RoomsUsecase) DeleteRoom(id int) error {

	if err := uc.repo.DeleteRoomByID(id); err != nil {
		return err
	}
	return nil
}
