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

func GetAllReviews(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviews, err := services.GetAllReviews(r.Context(), dbConnection)
		if err != nil {
			http.Error(w, "Unable to get all reviews", http.StatusServiceUnavailable)
			log.Println("Error getting reviews:", err.Error())
			return
		}

		var reviewDTOs []models.ReviewDTO
		for _, review := range reviews {
			reviewDTOs = append(reviewDTOs, review.ToDTO())
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, reviewDTOs)
	}
}

func GetReviewByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}
		review, err := services.GetReviewByID(r.Context(), dbConnection, reviewID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Review not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to get review", http.StatusServiceUnavailable)
			log.Println("Error getting review:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, review.ToDTO())
	}
}

func CreateReview(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reviewDTO models.ReviewDTO
		err := json.NewDecoder(r.Body).Decode(&reviewDTO)
		if err != nil {
			http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
			return
		}
		err = validator.Struct(reviewDTO)
		if err != nil {
			http.Error(w, models.ReviewValidationError, http.StatusBadRequest)
			return
		}
		newReview, err := reviewDTO.ToModel()
		if err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		err = services.CreateReview(r.Context(), dbConnection, &newReview)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "Unable to create review", http.StatusServiceUnavailable)
			log.Println("Error creating review:", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		returnJSON(w, newReview.ToDTO())
	}
}

func UpdateReviewByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}
		var reviewDTO models.ReviewDTO
		err = json.NewDecoder(r.Body).Decode(&reviewDTO)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		reviewDTO.BookingID = reviewID
		err = validator.Struct(reviewDTO)
		if err != nil {
			http.Error(w, models.ReviewValidationError, http.StatusBadRequest)
			return
		}
		updatedReview, err := reviewDTO.ToModel()
		if err != nil {
			http.Error(w, "Invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		status, err := services.UpdateReviewByID(r.Context(), dbConnection, &updatedReview)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Review not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to update review", http.StatusServiceUnavailable)
			log.Println("Error updating review:", err.Error())
			return
		}
		w.WriteHeader(status)
		returnJSON(w, updatedReview.ToDTO())
	}
}

func PatchReviewByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}
		var patch models.ReviewPatch
		err = json.NewDecoder(r.Body).Decode(&patch)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(patch)
		if err != nil {
			http.Error(w, "Review patch data is invalid", http.StatusBadRequest)
			return
		}
		err = services.PatchReviewByID(r.Context(), dbConnection, reviewID, patch)
		if err != nil {
			if errors.As(err, &models.ValidationError{}) {
				http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err == pgx.ErrNoRows {
				http.Error(w, "Review not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to patch review", http.StatusServiceUnavailable)
			log.Println("Error patching review:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteReviewByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reviewID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}
		err = services.DeleteReviewByID(r.Context(), dbConnection, reviewID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Review not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to delete review", http.StatusServiceUnavailable)
			log.Println("Error deleting review:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
