package main

import (
	"context"
	"example/handlers"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

func getDBConnStr() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func setupRoutes(mux *http.ServeMux, conn *pgx.Conn, validator *validator.Validate) {
	mux.HandleFunc("GET /", helloWorld)

	// customer
	mux.HandleFunc("GET /customers", handlers.GetAllCustomers(conn))
	mux.HandleFunc("GET /customers/{id}", handlers.GetCustomerByID(conn))
	mux.HandleFunc("POST /customers", handlers.CreateCustomer(conn, validator))
	mux.HandleFunc("PUT /customers/{id}", handlers.UpdateCustomerByID(conn, validator))
	mux.HandleFunc("PATCH /customers/{id}", handlers.PatchCustomerByID(conn, validator))
	mux.HandleFunc("DELETE /customers/{id}", handlers.DeleteCustomerByID(conn))

	// Booking
	mux.HandleFunc("GET /bookings", handlers.GetAllBookings(conn))
	mux.HandleFunc("GET /bookings/{id}", handlers.GetBookingByID(conn))
	mux.HandleFunc("POST /bookings", handlers.CreateBooking(conn, validator))
	mux.HandleFunc("PUT /bookings/{id}", handlers.UpdateBookingByID(conn, validator))
	mux.HandleFunc("PATCH /bookings/{id}", handlers.PatchBookingByID(conn, validator))
	mux.HandleFunc("DELETE /bookings/{id}", handlers.DeleteBookingByID(conn))

	// Review
	mux.HandleFunc("GET /reviews", handlers.GetAllReviews(conn))
	mux.HandleFunc("GET /reviews/{id}", handlers.GetReviewByID(conn))
	mux.HandleFunc("POST /reviews", handlers.CreateReview(conn, validator))
	mux.HandleFunc("PUT /reviews/{id}", handlers.UpdateReviewByID(conn, validator))
	mux.HandleFunc("PATCH /reviews/{id}", handlers.PatchReviewByID(conn, validator))
	mux.HandleFunc("DELETE /reviews/{id}", handlers.DeleteReviewByID(conn))

	// Room
	mux.HandleFunc("GET /rooms", handlers.GetAllRooms(conn))
	mux.HandleFunc("GET /rooms/{id}", handlers.GetRoomByID(conn))
	mux.HandleFunc("POST /rooms", handlers.CreateRoom(conn, validator))
	mux.HandleFunc("PUT /rooms/{id}", handlers.UpdateRoomByID(conn, validator))
	mux.HandleFunc("PATCH /rooms/{id}", handlers.PatchRoomByID(conn, validator))
	mux.HandleFunc("DELETE /rooms/{id}", handlers.DeleteRoomByID(conn))
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, getDBConnStr())
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer conn.Close(ctx)

	validator := validator.New()
	mux := http.NewServeMux()
	setupRoutes(mux, conn, validator)
	fmt.Printf("Server listening on http://%s:%s\n", os.Getenv("SERVER_HOST"), os.Getenv("SERVER_PORT"))
	http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), mux)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World"))
}
