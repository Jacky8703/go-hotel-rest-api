package services

import (
	"context"
	"example/dal"
	"example/models"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func GetAllReviews(ctx context.Context, conn *pgx.Conn) ([]models.Review, error) {
	return dal.GetAllReviews(ctx, conn)
}

func GetReviewByID(ctx context.Context, conn *pgx.Conn, reviewID int) (*models.Review, error) {
	return dal.GetReviewByID(ctx, conn, reviewID)
}

func CreateReview(ctx context.Context, conn *pgx.Conn, review *models.Review) error {
	err := validateReview(ctx, conn, review, true)
	if err != nil {
		return err
	}
	return dal.CreateReview(ctx, conn, review)
}

func UpdateReviewByID(ctx context.Context, conn *pgx.Conn, review *models.Review) (int, error) {
	err := validateReview(ctx, conn, review, false)
	if err != nil {
		return 0, err
	}
	_, err = dal.GetReviewByID(ctx, conn, review.BookingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusCreated, dal.CreateReview(ctx, conn, review)
		}
		return 0, err
	}
	return http.StatusOK, dal.UpdateReviewByID(ctx, conn, review)
}

func PatchReviewByID(ctx context.Context, conn *pgx.Conn, reviewID int, patch models.ReviewPatch) error {
	// first check that the patch is valid
	oldReview, err := dal.GetReviewByID(ctx, conn, reviewID)
	if err != nil {
		return err
	}

	if patch.BookingID != nil {
		oldReview.BookingID = *patch.BookingID
	}
	if patch.Comment != nil {
		oldReview.Comment = *patch.Comment
	}
	if patch.Rating != nil {
		oldReview.Rating = *patch.Rating
	}
	if patch.Date != nil {
		date, err := time.Parse("2006-01-02", *patch.Date)
		if err != nil {
			return models.ValidationError{Message: "date must be in YYYY-MM-DD format"}
		}
		oldReview.Date = date
	}

	err = validateReview(ctx, conn, oldReview, false)
	if err != nil {
		return err
	}

	return dal.PatchReviewByID(ctx, conn, reviewID, patch)
}

func DeleteReviewByID(ctx context.Context, conn *pgx.Conn, reviewID int) error {
	return dal.DeleteReviewByID(ctx, conn, reviewID)
}

func validateReview(ctx context.Context, conn *pgx.Conn, review *models.Review, new bool) error {
	booking, err := dal.GetBookingByID(ctx, conn, review.BookingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.ValidationError{Message: "booking does not exist"}
		}
		return err
	}
	if review.Date.Before(booking.StartDate) {
		return models.ValidationError{Message: "review date must be after booking start date"}
	}
	// check if the customer associated with the booking has already written a review, only if it wants to create another one
	if new {
		allReviews, err := dal.GetAllReviews(ctx, conn)
		if err != nil {
			return err
		}
		for _, r := range allReviews {
			if r.BookingID == review.BookingID {
				return models.ValidationError{Message: "customer has already written a review for this booking"}
			}
		}
	}
	return nil
}
