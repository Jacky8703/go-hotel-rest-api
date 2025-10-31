package services

import (
	"context"
	"errors"
	"example/dal"
	"example/models"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func GetAllBookings(ctx context.Context, conn *pgx.Conn) ([]models.Booking, error) {
	return dal.GetAllBookings(ctx, conn)
}

func GetBookingByID(ctx context.Context, conn *pgx.Conn, bookingID int) (*models.Booking, error) {
	return dal.GetBookingByID(ctx, conn, bookingID)
}

func CreateBooking(ctx context.Context, conn *pgx.Conn, booking *models.Booking) error {
	err := validateBooking(ctx, conn, booking)
	if err != nil {
		return err
	}
	return dal.CreateBooking(ctx, conn, booking)
}

func UpdateBookingByID(ctx context.Context, conn *pgx.Conn, booking *models.Booking) (int, error) {
	err := validateBooking(ctx, conn, booking)
	if err != nil {
		return 0, err
	}
	_, err = dal.GetBookingByID(ctx, conn, booking.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return http.StatusCreated, dal.CreateBooking(ctx, conn, booking)
		}
		return 0, err
	}
	return http.StatusOK, dal.UpdateBookingByID(ctx, conn, booking)
}

func PatchBookingByID(ctx context.Context, conn *pgx.Conn, bookingID int, patch models.BookingPatch) error {
	// first check that the patch is valid
	oldBooking, err := dal.GetBookingByID(ctx, conn, bookingID)
	if err != nil {
		return err
	}

	if patch.Code != nil {
		oldBooking.Code = *patch.Code
	}
	if patch.CustomerID != nil {
		oldBooking.CustomerID = *patch.CustomerID
	}
	if patch.RoomID != nil {
		oldBooking.RoomID = *patch.RoomID
	}
	if patch.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *patch.StartDate)
		if err != nil {
			return models.ValidationError{Message: "start date must be in YYYY-MM-DD format"}
		}
		oldBooking.StartDate = startDate
	}
	if patch.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *patch.EndDate)
		if err != nil {
			return models.ValidationError{Message: "end date must be in YYYY-MM-DD format"}
		}
		oldBooking.EndDate = endDate
	}
	err = validateBooking(ctx, conn, oldBooking)
	if err != nil {
		return err
	}

	return dal.PatchBookingByID(ctx, conn, bookingID, patch)
}

func DeleteBookingByID(ctx context.Context, conn *pgx.Conn, bookingID int) error {
	return dal.DeleteBookingByID(ctx, conn, bookingID)
}

func validateBooking(ctx context.Context, conn *pgx.Conn, booking *models.Booking) error {
	if booking.StartDate.After(booking.EndDate) {
		return models.ValidationError{Message: "start date must be before end date"}
	}
	// cannot create or update bookings in the past
	if booking.StartDate.Before(time.Now()) {
		return models.ValidationError{Message: "start date must be in the future"}
	}
	startYear, startMonth, startDay := booking.StartDate.Date()
	endYear, endMonth, endDay := booking.EndDate.Date()
	if startYear == endYear && startMonth == endMonth && startDay == endDay {
		return models.ValidationError{Message: "start date and end date cannot be the same"}
	}

	_, err := dal.GetCustomerByID(ctx, conn, booking.CustomerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ValidationError{Message: "customer does not exist"}
		}
		return err
	}
	_, err = dal.GetRoomByID(ctx, conn, booking.RoomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ValidationError{Message: "room does not exist"}
		}
		return err
	}

	// check for overlapping bookings for the same room
	allBookings, err := dal.GetAllBookings(ctx, conn)
	if err != nil {
		return err
	}
	for _, b := range allBookings {
		// different ID for update case
		if b.Code == booking.Code && b.ID != booking.ID {
			return models.ValidationError{Message: "booking code already exists"}
		}
		if booking.RoomID == b.RoomID && booking.ID != b.ID {
			// check for date overlap
			if booking.StartDate.Before(b.EndDate) && booking.EndDate.After(b.StartDate) {
				return models.ValidationError{Message: "booking dates overlap with an existing booking for the same room"}
			}
		}
	}
	return nil
}
