package dal

import (
	"context"
	"example/models"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
)

func GetAllBookings(ctx context.Context, conn *pgx.Conn) ([]models.Booking, error) {
	rows, _ := conn.Query(ctx, "SELECT id, code, customer_id, room_id, start_date, end_date FROM booking")
	defer rows.Close()
	var bookings []models.Booking
	for rows.Next() {
		var booking models.Booking
		err := rows.Scan(&booking.ID, &booking.Code, &booking.CustomerID, &booking.RoomID, &booking.StartDate, &booking.EndDate)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return bookings, nil
}

func GetBookingByID(ctx context.Context, conn *pgx.Conn, bookingID int) (*models.Booking, error) {
	row := conn.QueryRow(ctx, "SELECT id, code, customer_id, room_id, start_date, end_date FROM booking WHERE id = $1", bookingID)
	var booking models.Booking
	err := row.Scan(&booking.ID, &booking.Code, &booking.CustomerID, &booking.RoomID, &booking.StartDate, &booking.EndDate)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func CreateBooking(ctx context.Context, conn *pgx.Conn, booking *models.Booking) error {
	row := conn.QueryRow(ctx, "INSERT INTO booking (code, customer_id, room_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5) RETURNING id", booking.Code, booking.CustomerID, booking.RoomID, booking.StartDate, booking.EndDate)
	err := row.Scan(&booking.ID)
	return err
}

func UpdateBookingByID(ctx context.Context, conn *pgx.Conn, booking *models.Booking) error {
	row := conn.QueryRow(ctx, "UPDATE booking SET code = $1, customer_id = $2, room_id = $3, start_date = $4, end_date = $5 WHERE id = $6 RETURNING code, customer_id, room_id, start_date, end_date", booking.Code, booking.CustomerID, booking.RoomID, booking.StartDate, booking.EndDate, booking.ID)
	err := row.Scan(&booking.Code, &booking.CustomerID, &booking.RoomID, &booking.StartDate, &booking.EndDate)
	return err
}

func PatchBookingByID(ctx context.Context, conn *pgx.Conn, bookingID int, patch models.BookingPatch) error {
	query, args := createPatchQuery("booking", patch, "id", bookingID)
	tag, err := conn.Exec(ctx, query, args...)
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func DeleteBookingByID(ctx context.Context, conn *pgx.Conn, bookingID int) error {
	tag, err := conn.Exec(ctx, "DELETE FROM booking WHERE id = $1", bookingID)
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}

func createPatchQuery(table string, patch models.Patch, idName string, id int) (string, []any) {
	var changes []string
	var args []any
	argIndex := 1
	mapping := patch.FromStructToDBAttr()
	v := reflect.ValueOf(patch) // struct
	t := reflect.TypeOf(patch)  // model

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)     // pointer value
		fieldType := t.Field(i) // model field

		// pointer not nil
		if field.Kind() == reflect.Pointer && !field.IsNil() {
			changes = append(changes, fmt.Sprintf("%s = $%d", mapping[fieldType.Name], argIndex))
			args = append(args, field.Elem().Interface())
			argIndex++
		}
	}
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", table, strings.Join(changes, ", "), idName, argIndex)
	args = append(args, id)
	return query, args
}
