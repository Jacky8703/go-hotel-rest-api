package models

import "time"

type BookingDTO struct {
	ID         int    `json:"id,omitempty"`
	Code       string `json:"code" validate:"required"`
	CustomerID int    `json:"customer_id" validate:"required"`
	RoomID     int    `json:"room_id" validate:"required"`
	StartDate  string `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate    string `json:"end_date" validate:"required,datetime=2006-01-02"`
}

type Booking struct {
	ID         int
	Code       string
	CustomerID int
	RoomID     int
	StartDate  time.Time
	EndDate    time.Time
}

type Patch interface {
	FromStructToDBAttr() map[string]string
}

type BookingPatch struct {
	Code       *string `json:"code,omitempty"`
	CustomerID *int    `json:"customer_id,omitempty"`
	RoomID     *int    `json:"room_id,omitempty"`
	StartDate  *string `json:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	EndDate    *string `json:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

const BookingValidationError = `Invalid booking data:
- String field 'code' is required
- Integer field 'customer_id' is required
- Integer field 'room_id' is required
- String field 'start_date' is required and must be in YYYY-MM-DD format
- String field 'end_date' is required and must be in YYYY-MM-DD format`

func (b *Booking) ToDTO() BookingDTO {
	return BookingDTO{
		ID:         b.ID,
		Code:       b.Code,
		CustomerID: b.CustomerID,
		RoomID:     b.RoomID,
		StartDate:  b.StartDate.Format("2006-01-02"),
		EndDate:    b.EndDate.Format("2006-01-02"),
	}
}

func (b *BookingDTO) ToModel() (Booking, error) {
	startDate, err := time.Parse("2006-01-02", b.StartDate)
	if err != nil {
		return Booking{}, err
	}
	endDate, err := time.Parse("2006-01-02", b.EndDate)
	if err != nil {
		return Booking{}, err
	}
	return Booking{
		ID:         b.ID,
		Code:       b.Code,
		CustomerID: b.CustomerID,
		RoomID:     b.RoomID,
		StartDate:  startDate,
		EndDate:    endDate,
	}, nil
}

func (p BookingPatch) FromStructToDBAttr() map[string]string {
	return map[string]string{
		"Code":       "code",
		"CustomerID": "customer_id",
		"RoomID":     "room_id",
		"StartDate":  "start_date",
		"EndDate":    "end_date",
	}
}
