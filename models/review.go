package models

import "time"

type ReviewDTO struct {
	BookingID int    `json:"booking_id" validate:"required"`
	Comment   string `json:"comment" validate:"required"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
	Date      string `json:"date" validate:"required,datetime=2006-01-02"`
}

type Review struct {
	BookingID int
	Comment   string
	Rating    int
	Date      time.Time
}

type ReviewPatch struct {
	BookingID *int    `json:"booking_id,omitempty"`
	Comment   *string `json:"comment,omitempty"`
	Rating    *int    `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Date      *string `json:"date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

const ReviewValidationError = `Invalid review data:
- Integer field 'booking_id' is required
- String field 'comment' is required
- Integer field 'rating' is required and must be between 1 and 5
- String field 'date' is required and must be in YYYY-MM-DD format`

func (r *Review) ToDTO() ReviewDTO {
	return ReviewDTO{
		BookingID: r.BookingID,
		Comment:   r.Comment,
		Rating:    r.Rating,
		Date:      r.Date.Format("2006-01-02"),
	}
}

func (r *ReviewDTO) ToModel() (Review, error) {
	date, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		return Review{}, err
	}
	return Review{
		BookingID: r.BookingID,
		Comment:   r.Comment,
		Rating:    r.Rating,
		Date:      date,
	}, nil
}

func (p ReviewPatch) FromStructToDBAttr() map[string]string {
	return map[string]string{
		"BookingID": "booking_id",
		"Comment":   "review_comment",
		"Rating":    "rating",
		"Date":      "review_date",
	}
}
