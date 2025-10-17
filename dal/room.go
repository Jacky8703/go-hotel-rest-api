package dal

import (
	"context"
	"example/models"

	"github.com/jackc/pgx/v5"
)

func GetAllRooms(ctx context.Context, conn *pgx.Conn) ([]models.Room, error) {
	rows, _ := conn.Query(ctx, "SELECT id, room_number, room_type, price, capacity FROM room")
	defer rows.Close()
	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.Number, &room.Type, &room.Price, &room.Capacity)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return rooms, nil
}

func GetRoomByID(ctx context.Context, conn *pgx.Conn, roomID int) (*models.Room, error) {
	row := conn.QueryRow(ctx, "SELECT id, room_number, room_type, price, capacity FROM room WHERE id = $1", roomID)
	var room models.Room
	err := row.Scan(&room.ID, &room.Number, &room.Type, &room.Price, &room.Capacity)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func CreateRoom(ctx context.Context, conn *pgx.Conn, room *models.Room) error {
	row := conn.QueryRow(ctx, "INSERT INTO room (room_number, room_type, price, capacity) VALUES ($1, $2, $3, $4) RETURNING id", room.Number, room.Type, room.Price, room.Capacity)
	err := row.Scan(&room.ID)
	return err
}

func UpdateRoomByID(ctx context.Context, conn *pgx.Conn, room *models.Room) error {
	row := conn.QueryRow(ctx, "UPDATE room SET room_number = $1, room_type = $2, price = $3, capacity = $4 WHERE id = $5 RETURNING room_number, room_type, price, capacity", room.Number, room.Type, room.Price, room.Capacity, room.ID)
	err := row.Scan(&room.Number, &room.Type, &room.Price, &room.Capacity)
	return err
}

func PatchRoomByID(ctx context.Context, conn *pgx.Conn, roomID int, patch models.RoomPatch) error {
	query, args := createPatchQuery("room", patch, roomID)
	tag, err := conn.Exec(ctx, query, args...)
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func DeleteRoomByID(ctx context.Context, conn *pgx.Conn, roomID int) error {
	tag, err := conn.Exec(ctx, "DELETE FROM room WHERE id = $1", roomID)
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
