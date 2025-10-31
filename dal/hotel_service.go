package dal

import (
	"context"
	"example/models"

	"github.com/jackc/pgx/v5"
)

func GetAllHotelServices(ctx context.Context, conn *pgx.Conn) ([]models.HotelService, error) {
	rows, _ := conn.Query(ctx, "SELECT id, service_type, description, duration FROM hotel_service")
	defer rows.Close()
	var services []models.HotelService
	for rows.Next() {
		var service models.HotelService
		err := rows.Scan(&service.ID, &service.Type, &service.Description, &service.Duration)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return services, nil
}

func GetHotelServiceByID(ctx context.Context, conn *pgx.Conn, serviceID int) (*models.HotelService, error) {
	row := conn.QueryRow(ctx, "SELECT id, service_type, description, duration FROM hotel_service WHERE id = $1", serviceID)
	var service models.HotelService
	err := row.Scan(&service.ID, &service.Type, &service.Description, &service.Duration)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func CreateHotelService(ctx context.Context, conn *pgx.Conn, service *models.HotelService) error {
	row := conn.QueryRow(ctx, "INSERT INTO hotel_service (service_type, description, duration) VALUES ($1, $2, $3) RETURNING id", service.Type, service.Description, service.Duration)
	err := row.Scan(&service.ID)
	return err
}

func UpdateHotelServiceByID(ctx context.Context, conn *pgx.Conn, service *models.HotelService) error {
	row := conn.QueryRow(ctx, "UPDATE hotel_service SET service_type = $1, description = $2, duration = $3 WHERE id = $4 RETURNING service_type, description, duration", service.Type, service.Description, service.Duration, service.ID)
	err := row.Scan(&service.Type, &service.Description, &service.Duration)
	return err
}

func PatchHotelServiceByID(ctx context.Context, conn *pgx.Conn, serviceID int, patch models.HotelServicePatch) error {
	query, args := createPatchQuery("hotel_service", patch, "id", serviceID)
	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func DeleteHotelServiceByID(ctx context.Context, conn *pgx.Conn, serviceID int) error {
	tag, err := conn.Exec(ctx, "DELETE FROM hotel_service WHERE id = $1", serviceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
