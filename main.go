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

func getDBConnStr(host string, dbName string) string {
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName)
}

func setupRoutes(mux *http.ServeMux, conn *pgx.Conn, validator *validator.Validate) {
	mux.HandleFunc("GET /", helloWorld)

	// Customers
	mux.HandleFunc("GET /customers", handlers.GetAllCustomers(conn))
	mux.HandleFunc("GET /customers/{id}", handlers.GetCustomerByID(conn))
	mux.HandleFunc("POST /customers", handlers.CreateCustomer(conn, validator))
	mux.HandleFunc("PUT /customers/{id}", handlers.UpdateCustomerByID(conn, validator))
	mux.HandleFunc("PATCH /customers/{id}", handlers.PatchCustomerByID(conn, validator))
	mux.HandleFunc("DELETE /customers/{id}", handlers.DeleteCustomerByID(conn))

	// Bookings
	mux.HandleFunc("GET /bookings", handlers.GetAllBookings(conn))
	mux.HandleFunc("GET /bookings/{id}", handlers.GetBookingByID(conn))
	mux.HandleFunc("POST /bookings", handlers.CreateBooking(conn, validator))
	mux.HandleFunc("PUT /bookings/{id}", handlers.UpdateBookingByID(conn, validator))
	mux.HandleFunc("PATCH /bookings/{id}", handlers.PatchBookingByID(conn, validator))
	mux.HandleFunc("DELETE /bookings/{id}", handlers.DeleteBookingByID(conn))

	// Reviews
	mux.HandleFunc("GET /reviews", handlers.GetAllReviews(conn))
	mux.HandleFunc("GET /reviews/{id}", handlers.GetReviewByID(conn))
	mux.HandleFunc("POST /reviews", handlers.CreateReview(conn, validator))
	mux.HandleFunc("PUT /reviews/{id}", handlers.UpdateReviewByID(conn, validator))
	mux.HandleFunc("PATCH /reviews/{id}", handlers.PatchReviewByID(conn, validator))
	mux.HandleFunc("DELETE /reviews/{id}", handlers.DeleteReviewByID(conn))

	// Rooms
	mux.HandleFunc("GET /rooms", handlers.GetAllRooms(conn))
	mux.HandleFunc("GET /rooms/{id}", handlers.GetRoomByID(conn))
	mux.HandleFunc("POST /rooms", handlers.CreateRoom(conn, validator))
	mux.HandleFunc("PUT /rooms/{id}", handlers.UpdateRoomByID(conn, validator))
	mux.HandleFunc("PATCH /rooms/{id}", handlers.PatchRoomByID(conn, validator))
	mux.HandleFunc("DELETE /rooms/{id}", handlers.DeleteRoomByID(conn))

	// Services
	mux.HandleFunc("GET /services", handlers.GetAllHotelServices(conn))
	mux.HandleFunc("GET /services/{id}", handlers.GetHotelServiceByID(conn))
	mux.HandleFunc("POST /services", handlers.CreateHotelService(conn, validator))
	mux.HandleFunc("PUT /services/{id}", handlers.UpdateHotelServiceByID(conn, validator))
	mux.HandleFunc("PATCH /services/{id}", handlers.PatchHotelServiceByID(conn, validator))
	mux.HandleFunc("DELETE /services/{id}", handlers.DeleteHotelServiceByID(conn))

	// Service Requests
	mux.HandleFunc("GET /service-requests", handlers.GetAllServiceRequests(conn))
	mux.HandleFunc("GET /service-requests/{id}", handlers.GetServiceRequestByID(conn))
	mux.HandleFunc("POST /service-requests", handlers.CreateServiceRequest(conn, validator))
	mux.HandleFunc("PUT /service-requests/{id}", handlers.UpdateServiceRequestByID(conn, validator))
	mux.HandleFunc("PATCH /service-requests/{id}", handlers.PatchServiceRequestByID(conn, validator))
	mux.HandleFunc("DELETE /service-requests/{id}", handlers.DeleteServiceRequestByID(conn))
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, getDBConnStr(os.Getenv("DB_HOST"), os.Getenv("DB_NAME")))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer conn.Close(ctx)

	val := validator.New()
	mux := http.NewServeMux()
	setupRoutes(mux, conn, val)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server listening on http://0.0.0.0:%s\n", port)
	err = http.ListenAndServe("0.0.0.0:"+port, mux)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World"))
}
