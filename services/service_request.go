package services

import (
	"context"
	"example/dal"
	"example/models"
	"time"

	"github.com/jackc/pgx/v5"
)

func GetAllServiceRequests(ctx context.Context, conn *pgx.Conn) ([]models.ServiceRequest, error) {
	return dal.GetAllServiceRequests(ctx, conn)
}

func GetServiceRequestByID(ctx context.Context, conn *pgx.Conn, requestID int) (*models.ServiceRequest, error) {
	return dal.GetServiceRequestByID(ctx, conn, requestID)
}

func CreateServiceRequest(ctx context.Context, conn *pgx.Conn, request *models.ServiceRequest) error {
	err := validateServiceRequest(ctx, conn, request)
	if err != nil {
		return err
	}
	return dal.CreateServiceRequest(ctx, conn, request)
}

func UpdateServiceRequestByID(ctx context.Context, conn *pgx.Conn, request *models.ServiceRequest) error {
	err := validateServiceRequest(ctx, conn, request)
	if err != nil {
		return err
	}
	return dal.UpdateServiceRequestByID(ctx, conn, request)
}

func PatchServiceRequestByID(ctx context.Context, conn *pgx.Conn, requestID int, patch models.ServiceRequestPatch) error {
	// first check that the patch is valid
	oldRequest, err := dal.GetServiceRequestByID(ctx, conn, requestID)
	if err != nil {
		return err
	}

	if patch.CustomerID != nil {
		oldRequest.CustomerID = *patch.CustomerID
	}
	if patch.ServiceID != nil {
		oldRequest.ServiceID = *patch.ServiceID
	}
	if patch.Date != nil {
		date, err := time.Parse("2006-01-02", *patch.Date)
		if err != nil {
			return models.ValidationError{Message: "service request date must be in YYYY-MM-DD format"}
		}
		oldRequest.Date = date
	}

	err = validateServiceRequest(ctx, conn, oldRequest)
	if err != nil {
		return err
	}

	return dal.PatchServiceRequestByID(ctx, conn, requestID, patch)
}

func DeleteServiceRequestByID(ctx context.Context, conn *pgx.Conn, requestID int) error {
	return dal.DeleteServiceRequestByID(ctx, conn, requestID)
}

func validateServiceRequest(ctx context.Context, conn *pgx.Conn, request *models.ServiceRequest) error {
	if request.Date.Before(time.Now()) {
		return models.ValidationError{Message: "service request date must be in the future"}
	}
	customer, err := dal.GetCustomerByID(ctx, conn, request.CustomerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.ValidationError{Message: "customer does not exist"}
		}
		return err
	}
	bookings, err := dal.GetAllBookings(ctx, conn)
	if err != nil {
		return err
	}
	if len(bookings) == 0 {
		return models.ValidationError{Message: "customer has no bookings"}
	}
	for _, booking := range bookings {
		if booking.CustomerID == customer.ID {
			if request.Date.After(booking.EndDate) || request.Date.Before(booking.StartDate) {
				return models.ValidationError{Message: "service request date must be within a booking period"}
			}
		}
	}
	_, err = dal.GetHotelServiceByID(ctx, conn, request.ServiceID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.ValidationError{Message: "service does not exist"}
		}
		return err
	}
	requests, err := dal.GetAllServiceRequests(ctx, conn)
	if err != nil {
		return err
	}
	for _, r := range requests {
		if r.CustomerID == request.CustomerID && r.ServiceID == request.ServiceID && r.Date.Equal(request.Date) && r.ID != request.ID {
			return models.ValidationError{Message: "duplicate service request"}
		}
	}

	return nil
}
