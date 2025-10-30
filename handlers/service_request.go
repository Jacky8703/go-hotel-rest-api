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

func GetAllServiceRequests(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requests, err := services.GetAllServiceRequests(r.Context(), dbConnection)
		if err != nil {
			http.Error(w, "Unable to get all service requests", http.StatusServiceUnavailable)
			log.Println("Error getting service requests:", err.Error())
			return
		}

		var requestDTOs []models.ServiceRequestDTO
		for _, request := range requests {
			requestDTOs = append(requestDTOs, request.ToDTO())
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, requestDTOs)
	}
}

func GetServiceRequestByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid service request ID", http.StatusBadRequest)
			return
		}
		request, err := services.GetServiceRequestByID(r.Context(), dbConnection, requestID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Service request not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to get service request", http.StatusServiceUnavailable)
			log.Println("Error getting service request:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, request.ToDTO())
	}
}

func CreateServiceRequest(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestDTO models.ServiceRequestDTO
		err := json.NewDecoder(r.Body).Decode(&requestDTO)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(requestDTO)
		if err != nil {
			http.Error(w, models.ServiceRequestValidationError, http.StatusBadRequest)
			return
		}
		request, err := requestDTO.ToModel()
		if err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		request.ID = -1 // ensure ID is invalid for creation
		err = services.CreateServiceRequest(r.Context(), dbConnection, &request)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Unable to create service request", http.StatusServiceUnavailable)
			log.Println("Error creating service request:", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		returnJSON(w, request.ToDTO())
	}
}

func UpdateServiceRequestByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid service request ID", http.StatusBadRequest)
			return
		}
		var requestDTO models.ServiceRequestDTO
		err = json.NewDecoder(r.Body).Decode(&requestDTO)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		requestDTO.ID = requestID
		err = validator.Struct(requestDTO)
		if err != nil {
			http.Error(w, models.ServiceRequestValidationError, http.StatusBadRequest)
			return
		}
		request, err := requestDTO.ToModel()
		if err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		status, err := services.UpdateServiceRequestByID(r.Context(), dbConnection, &request)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Service request not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to update service request", http.StatusServiceUnavailable)
			log.Println("Error updating service request:", err.Error())
			return
		}
		w.WriteHeader(status)
		returnJSON(w, request.ToDTO())
	}
}

func PatchServiceRequestByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid service request ID", http.StatusBadRequest)
			return
		}
		var patch models.ServiceRequestPatch
		err = json.NewDecoder(r.Body).Decode(&patch)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(patch)
		if err != nil {
			http.Error(w, "Service Request patch data is invalid", http.StatusBadRequest)
			return
		}
		err = services.PatchServiceRequestByID(r.Context(), dbConnection, requestID, patch)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Service request not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to patch service request", http.StatusServiceUnavailable)
			log.Println("Error patching service request:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteServiceRequestByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid service request ID", http.StatusBadRequest)
			return
		}
		err = services.DeleteServiceRequestByID(r.Context(), dbConnection, requestID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Service request not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to delete service request", http.StatusServiceUnavailable)
			log.Println("Error deleting service request:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
