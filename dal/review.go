package dal

import (
	"context"
	"example/models"

	"github.com/jackc/pgx/v5"
)

func GetAllReviews(ctx context.Context, conn *pgx.Conn) ([]models.Review, error) {
	rows, _ := conn.Query(ctx, "SELECT booking_id, review_comment, rating, review_date FROM review")
	defer rows.Close()
	var reviews []models.Review
	for rows.Next() {
		var review models.Review
		err := rows.Scan(&review.BookingID, &review.Comment, &review.Rating, &review.Date)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return reviews, nil
}

func GetReviewByID(ctx context.Context, conn *pgx.Conn, reviewID int) (*models.Review, error) {
	row := conn.QueryRow(ctx, "SELECT booking_id, review_comment, rating, review_date FROM review WHERE booking_id = $1", reviewID)
	var review models.Review
	err := row.Scan(&review.BookingID, &review.Comment, &review.Rating, &review.Date)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func CreateReview(ctx context.Context, conn *pgx.Conn, review *models.Review) error {
	row := conn.QueryRow(ctx, "INSERT INTO review (booking_id, review_comment, rating, review_date) VALUES ($1, $2, $3, $4) RETURNING booking_id", review.BookingID, review.Comment, review.Rating, review.Date)
	err := row.Scan(&review.BookingID)
	return err
}

func UpdateReviewByID(ctx context.Context, conn *pgx.Conn, review *models.Review) error {
	row := conn.QueryRow(ctx, "UPDATE review SET review_comment = $1, rating = $2, review_date = $3 WHERE booking_id = $4 RETURNING review_comment, rating, review_date", review.Comment, review.Rating, review.Date, review.BookingID)
	err := row.Scan(&review.Comment, &review.Rating, &review.Date)
	return err
}

func PatchReviewByID(ctx context.Context, conn *pgx.Conn, reviewID int, patch models.ReviewPatch) error {
	query, args := createPatchQuery("review", patch, "booking_id", reviewID)
	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func DeleteReviewByID(ctx context.Context, conn *pgx.Conn, reviewID int) error {
	tag, err := conn.Exec(ctx, "DELETE FROM review WHERE booking_id = $1", reviewID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
