package dal

import (
	"context"
	"example/models"

	"github.com/jackc/pgx/v5"
)

func GetAllServiceRequests(ctx context.Context, conn *pgx.Conn) ([]models.ServiceRequest, error) {
	rows, _ := conn.Query(ctx, "SELECT id, customer_id, service_id, service_date FROM service_request")
	defer rows.Close()
	var requests []models.ServiceRequest
	for rows.Next() {
		var request models.ServiceRequest
		err := rows.Scan(&request.ID, &request.CustomerID, &request.ServiceID, &request.Date)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return requests, nil
}

func GetServiceRequestByID(ctx context.Context, conn *pgx.Conn, requestID int) (*models.ServiceRequest, error) {
	row := conn.QueryRow(ctx, "SELECT id, customer_id, service_id, service_date FROM service_request WHERE id = $1", requestID)
	var request models.ServiceRequest
	err := row.Scan(&request.ID, &request.CustomerID, &request.ServiceID, &request.Date)
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func CreateServiceRequest(ctx context.Context, conn *pgx.Conn, request *models.ServiceRequest) error {
	row := conn.QueryRow(ctx, "INSERT INTO service_request (customer_id, service_id, service_date) VALUES ($1, $2, $3) RETURNING id", request.CustomerID, request.ServiceID, request.Date)
	err := row.Scan(&request.ID)
	return err
}

func UpdateServiceRequestByID(ctx context.Context, conn *pgx.Conn, request *models.ServiceRequest) error {
	row := conn.QueryRow(ctx, "UPDATE service_request SET customer_id = $1, service_id = $2, service_date = $3 WHERE id = $4 RETURNING customer_id, service_id, service_date", request.CustomerID, request.ServiceID, request.Date, request.ID)
	err := row.Scan(&request.CustomerID, &request.ServiceID, &request.Date)
	return err
}

func PatchServiceRequestByID(ctx context.Context, conn *pgx.Conn, requestID int, patch models.ServiceRequestPatch) error {
	query, args := createPatchQuery("service_request", patch, "id", requestID)
	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func DeleteServiceRequestByID(ctx context.Context, conn *pgx.Conn, requestID int) error {
	tag, err := conn.Exec(ctx, "DELETE FROM service_request WHERE id = $1", requestID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
