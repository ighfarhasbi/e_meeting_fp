package repository

import (
	"database/sql"
	"e_meeting/internal/entity"
	"e_meeting/internal/models/response"
	"fmt"
)

type DBRoomsRepository struct {
	DB *sql.DB
}

func NewDBRoomsRepository(db *sql.DB) *DBRoomsRepository {
	return &DBRoomsRepository{
		DB: db,
	}
}

func (r *DBRoomsRepository) GetAllRooms(roomName string, roomType string, capacity int, pageSize int, offset int) ([]entity.Rooms, error) {
	rows, err := r.DB.Query(`
    	SELECT rooms_id, name, type, price_perhour, capacity, img_path, created_at, updated_at 
    	FROM rooms
    	WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
    		AND ($2 = '' OR type = $2::room_type)
      		AND ($3 = 0 OR capacity >= $3)
    	ORDER BY rooms_id
    	LIMIT $4 OFFSET $5
		`, roomName, roomType, capacity, pageSize, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return []entity.Rooms{}, nil
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer rows.Close()

	var roomsList []entity.Rooms
	for rows.Next() {
		var room entity.Rooms
		if err := rows.Scan(&room.ID, &room.Name, &room.Type, &room.PricePerHour, &room.Capacity, &room.ImgUrl, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalServer, err)
		}
		roomsList = append(roomsList, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServer, err)
	}

	return roomsList, nil
}
func (r *DBRoomsRepository) CountTotalRooms(roomName string, roomType string, capacity int) (int, error) {
	var total int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM rooms
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
			AND ($2 = '' OR type = $2::room_type)
			AND ($3 = 0 OR capacity >= $3)
		`, roomName, roomType, capacity).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	return total, nil
}

func (r *DBRoomsRepository) CreateRoom(room *entity.Rooms) error {
	_, err := r.DB.Exec("INSERT INTO rooms (name, type, price_perhour, capacity, img_path) VALUES ($1, $2, $3, $4, $5)",
		room.Name, room.Type, room.PricePerHour, room.Capacity, room.ImgUrl)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalServer, err)
	}
	return nil
}

func (r *DBRoomsRepository) UpdateRoomByID(room *entity.Rooms) error {
	result, err := r.DB.Exec("UPDATE rooms SET name = $1, type = $2, price_perhour = $3, capacity = $4, img_path = $5 WHERE rooms_id = $6",
		room.Name, room.Type, room.PricePerHour, room.Capacity, room.ImgUrl, room.ID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalServer, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalServer, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %v", ErrRoomNotFound, err) // Room not found
	}
	return nil
}

func (r *DBRoomsRepository) DeleteRoomByID(id int) error {
	result, err := r.DB.Exec("DELETE FROM rooms WHERE rooms_id = $1", id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalServer, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalServer, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: %v", ErrRoomNotFound, err) // Room not found
	}
	return nil
}

func (r *DBRoomsRepository) GetRoomsByIDAndDate(id int, date string) (*response.RoomsSchedules, error) {
	// get room_name by id
	var roomName string
	err := r.DB.QueryRow("SELECT name FROM rooms WHERE rooms_id = $1", id).Scan(&roomName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRoomNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	// ambil data room dari database yang difilter berdasarkan id dan status
	rows, err := r.DB.Query(`select t.status, dt.start_time, dt.end_time 
	from transactions t
	join detail_transaction dt on t.tx_id = dt.tx_id
	where dt.rooms_id = $1 and t.status != 'canceled' AND dt.start_time >= $2::date AND dt.start_time < ($2::date + INTERVAL '1 day')`,
		id, date)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	defer rows.Close()

	var sch []response.Schedules
	for rows.Next() {
		var s response.Schedules
		if err := rows.Scan(&s.Status, &s.StartTime, &s.EndTime); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalServer, err)
		}
		sch = append(sch, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServer, err)
	}
	resultSch := response.RoomsSchedules{
		RoomName:    roomName,
		Schedules:   sch,
		TotalBooked: len(sch),
	}
	return &resultSch, nil
}

// func (r *DBRoomsRepository) GetRoomsSchedule()
