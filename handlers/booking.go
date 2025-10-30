package handlers

import (
	"encoding/json"
	"errors"
	"example/models"
	"example/services"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

func GetAllBookings(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookings, err := services.GetAllBookings(r.Context(), dbConnection)
		if err != nil {
			http.Error(w, "Unable to get all bookings", http.StatusServiceUnavailable)
			log.Println("Error getting bookings:", err.Error())
			return
		}

		var bookingDTOs []models.BookingDTO
		for _, booking := range bookings {
			bookingDTOs = append(bookingDTOs, booking.ToDTO())
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, bookingDTOs)
	}
}

func GetBookingByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookingID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid booking ID", http.StatusBadRequest)
			return
		}
		booking, err := services.GetBookingByID(r.Context(), dbConnection, bookingID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Booking not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to get booking", http.StatusServiceUnavailable)
			log.Println("Error getting booking:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, booking.ToDTO())
	}
}

func CreateBooking(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var bookingDTO models.BookingDTO
		err := json.NewDecoder(r.Body).Decode(&bookingDTO)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(bookingDTO)
		if err != nil {
			http.Error(w, models.BookingValidationError, http.StatusBadRequest)
			return
		}
		newBooking, err := bookingDTO.ToModel()
		if err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		newBooking.ID = -1 // ensure ID is invalid for creation
		err = services.CreateBooking(r.Context(), dbConnection, &newBooking)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Unable to create booking", http.StatusServiceUnavailable)
			log.Println("Error creating booking:", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		returnJSON(w, newBooking.ToDTO())
	}
}

func UpdateBookingByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookingID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid booking ID", http.StatusBadRequest)
			return
		}
		var bookingDTO models.BookingDTO
		err = json.NewDecoder(r.Body).Decode(&bookingDTO)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		bookingDTO.ID = bookingID
		err = validator.Struct(bookingDTO)
		if err != nil {
			http.Error(w, models.BookingValidationError, http.StatusBadRequest)
			return
		}
		updatedBooking, err := bookingDTO.ToModel()
		if err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		err = services.UpdateBookingByID(r.Context(), dbConnection, &updatedBooking)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Booking not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to update booking", http.StatusServiceUnavailable)
			log.Println("Error updating booking:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, updatedBooking.ToDTO())
	}
}

func PatchBookingByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookingID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid booking ID", http.StatusBadRequest)
			return
		}
		var patch models.BookingPatch
		err = json.NewDecoder(r.Body).Decode(&patch)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(patch)
		if err != nil {
			http.Error(w, "Booking patch data is invalid", http.StatusBadRequest)
			return
		}
		err = services.PatchBookingByID(r.Context(), dbConnection, bookingID, patch)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Booking not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to patch booking", http.StatusServiceUnavailable)
			log.Println("Error patching booking:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteBookingByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bookingID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid booking ID", http.StatusBadRequest)
			return
		}
		err = services.DeleteBookingByID(r.Context(), dbConnection, bookingID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Booking not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to delete booking", http.StatusServiceUnavailable)
			log.Println("Error deleting booking:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
