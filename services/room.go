package services

import (
	"context"
	"example/dal"
	"example/models"

	"github.com/jackc/pgx/v5"
)

func GetAllRooms(ctx context.Context, conn *pgx.Conn) ([]models.Room, error) {
	return dal.GetAllRooms(ctx, conn)
}

func GetRoomByID(ctx context.Context, conn *pgx.Conn, roomID int) (*models.Room, error) {
	return dal.GetRoomByID(ctx, conn, roomID)
}

func CreateRoom(ctx context.Context, conn *pgx.Conn, room *models.Room) error {
	return dal.CreateRoom(ctx, conn, room)
}

func UpdateRoomByID(ctx context.Context, conn *pgx.Conn, room *models.Room) error {
	return dal.UpdateRoomByID(ctx, conn, room)
}

func PatchRoomByID(ctx context.Context, conn *pgx.Conn, roomID int, patch models.RoomPatch) error {
	return dal.PatchRoomByID(ctx, conn, roomID, patch)
}

func DeleteRoomByID(ctx context.Context, conn *pgx.Conn, roomID int) error {
	return dal.DeleteRoomByID(ctx, conn, roomID)
}
