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

func GetAllHotelServices(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		services, err := services.GetAllHotelServices(r.Context(), dbConnection)
		if err != nil {
			http.Error(w, "Unable to get all hotel services", http.StatusServiceUnavailable)
			log.Println("Error getting hotel services:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, services)
	}
}

func GetHotelServiceByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid hotel service ID", http.StatusBadRequest)
			return
		}
		service, err := services.GetHotelServiceByID(r.Context(), dbConnection, serviceID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Hotel service not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to get hotel service", http.StatusServiceUnavailable)
			log.Println("Error getting hotel service:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, service)
	}
}

func CreateHotelService(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var service models.HotelService
		err := json.NewDecoder(r.Body).Decode(&service)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(service)
		if err != nil {
			http.Error(w, models.HotelServiceValidationError, http.StatusBadRequest)
			return
		}
		service.ID = -1 // ensure ID is invalid for creation
		err = services.CreateHotelService(r.Context(), dbConnection, &service)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Unable to create hotel service", http.StatusServiceUnavailable)
			log.Println("Error creating hotel service:", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		returnJSON(w, service)
	}
}

func UpdateHotelServiceByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid hotel service ID", http.StatusBadRequest)
			return
		}
		var service models.HotelService
		err = json.NewDecoder(r.Body).Decode(&service)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		service.ID = serviceID
		err = validator.Struct(service)
		if err != nil {
			http.Error(w, models.HotelServiceValidationError, http.StatusBadRequest)
			return
		}
		err = services.UpdateHotelServiceByID(r.Context(), dbConnection, &service)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Hotel service not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to update hotel service", http.StatusServiceUnavailable)
			log.Println("Error updating hotel service:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, service)
	}
}

func PatchHotelServiceByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid hotel service ID", http.StatusBadRequest)
			return
		}
		var patch models.HotelServicePatch
		err = json.NewDecoder(r.Body).Decode(&patch)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(patch)
		if err != nil {
			http.Error(w, "Hotel service patch data is invalid", http.StatusBadRequest)
			return
		}
		err = services.PatchHotelServiceByID(r.Context(), dbConnection, serviceID, patch)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Hotel service not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to patch hotel service", http.StatusServiceUnavailable)
			log.Println("Error patching hotel service:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteHotelServiceByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid hotel service ID", http.StatusBadRequest)
			return
		}
		err = services.DeleteHotelServiceByID(r.Context(), dbConnection, serviceID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Hotel service not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to delete hotel service", http.StatusServiceUnavailable)
			log.Println("Error deleting hotel service:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
