package repository

import "errors"

var (
	ErrRoomNotFound       = errors.New("room not found")
	ErrRoomAlreadyExists  = errors.New("room already exists")
	ErrFailedToCreateRoom = errors.New("failed to create room")
	ErrFailedToUpdateRoom = errors.New("failed to update room")
	ErrFailedToDeleteRoom = errors.New("failed to delete room")
	ErrInvalidRoomID      = errors.New("invalid room ID")
	ErrNoRoomsAvailable   = errors.New("no rooms available")
	ErrDatabase           = errors.New("database error")
	ErrInternalServer     = errors.New("internal server error")
)
