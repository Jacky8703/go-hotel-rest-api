package main

import (
	"bytes"
	"context"
	"encoding/json"
	"example/models"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

var (
	testDBName  = "testdb"
	schemaPath  = "schema.sql"
	conn        *pgx.Conn
	client      = &http.Client{}
	baseURI     string
	roomURI     string
	customerURI string
	bookingURI  string
	reviewURI   string
	serviceURI  string
	requestURI  string
	sampleRoom  = models.Room{
		Number:   101,
		Type:     "basic",
		Price:    100,
		Capacity: 2,
	}
	sampleCustomer = models.Customer{
		CF:    "TESTCF12345",
		Name:  "Testino",
		Age:   30,
		Email: "testcustomer@example.com",
	}
	sampleBookingDTO = models.BookingDTO{
		Code:       "TESTBOOK123",
		CustomerID: -1,
		RoomID:     -1,
		StartDate:  time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		EndDate:    time.Now().AddDate(0, 0, 8).Format("2006-01-02"),
	}
	sampleReviewDTO = models.ReviewDTO{
		BookingID: -1,
		Comment:   "comment",
		Rating:    3,
		Date:      time.Now().AddDate(0, 0, 9).Format("2006-01-02"),
	}
	sampleService = models.HotelService{
		Type:        "room_service",
		Description: "Sample description",
		Duration:    30,
	}
	sampleServiceRequestDTO = models.ServiceRequestDTO{
		CustomerID: -1,
		ServiceID:  -1,
		Date:       time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
	}
)

func TestMain(m *testing.M) {
	// connect to the admin database and create a new test database
	ctx := context.Background()
	adminConnStr := getDBConnStr("localhost", os.Getenv("DB_NAME"))
	adminConn, err := pgx.Connect(ctx, adminConnStr)
	if err != nil {
		fmt.Println("Unable to connect to admin database:", err)
		os.Exit(1)
	}
	defer adminConn.Close(ctx)

	_, err = adminConn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	if err != nil {
		fmt.Println("Unable to drop test database:", err)
		os.Exit(1)
	}
	_, err = adminConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		fmt.Println("Unable to create test database:", err)
		os.Exit(1)
	}
	conn, err = pgx.Connect(ctx, getDBConnStr("localhost", testDBName))
	if err != nil {
		fmt.Println("Unable to connect to test database:", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// populate the database schema
	sql, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Printf("Unable to read %s: %v\n", schemaPath, err)
		os.Exit(1)
	}
	_, err = conn.Exec(ctx, string(sql))
	if err != nil {
		fmt.Printf("Unable to execute %s: %v\n", schemaPath, err)
		os.Exit(1)
	}

	val := validator.New()
	mux := http.NewServeMux()
	setupRoutes(mux, conn, val)
	testServer := httptest.NewServer(mux)
	baseURI = testServer.URL
	roomURI = baseURI + "/rooms"
	customerURI = baseURI + "/customers"
	bookingURI = baseURI + "/bookings"
	reviewURI = baseURI + "/reviews"
	serviceURI = baseURI + "/services"
	requestURI = baseURI + "/service-requests"
	defer testServer.Close()

	code := m.Run()

	os.Exit(code)
}

func TestHelloWorld(t *testing.T) {
	resp, err := http.Get(baseURI + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// truncate all tables
func resetDatabase(t *testing.T) {
	ctx := context.Background()
	_, err := conn.Exec(ctx, "TRUNCATE TABLE customer, booking, review, service_request, hotel_service, room RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to truncate tables: %v", err)
}

// helper function to make HTTP requests, returns response and body
func makeRequest(t *testing.T, method, path string, body any) (*http.Response, []byte) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body: %v", err)
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(t, err, "Failed to create HTTP request: %v", err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to execute HTTP request: %v", err)

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err, "Failed to read response body: %v", err)
	return resp, respBody
}

// helper function to create a sample entity in the database
func createSample[T any](t *testing.T, uri string, model T) T {
	resp, body := makeRequest(t, http.MethodPost, uri, model)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create sample: %s", string(body))
	var createdModel T
	err := json.Unmarshal(body, &createdModel)
	require.NoError(t, err, "Failed to unmarshal created model: %v", err)
	return createdModel
}

func TestCustomerEndpoints(t *testing.T) {
	t.Run("POST/customers", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)
		require.Equal(t, sampleCustomer.CF, newCustomer.CF)
	})
	t.Run("GET/customers", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)

		resp, body := makeRequest(t, http.MethodGet, customerURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var customers []models.Customer
		err := json.Unmarshal(body, &customers)
		require.NoError(t, err)
		require.Contains(t, customers, newCustomer)
	})
	t.Run("GET/customers/{id}", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", customerURI, newCustomer.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var customer models.Customer
		err := json.Unmarshal(body, &customer)
		require.NoError(t, err)
		require.Equal(t, newCustomer, customer)
	})
	t.Run("PUT/customers/{id} - update", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)
		newCustomer.Name = "UpdatedName"

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", customerURI, newCustomer.ID), newCustomer)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var customer models.Customer
		err := json.Unmarshal(body, &customer)
		require.NoError(t, err)
		require.Equal(t, newCustomer, customer)
	})
	t.Run("PUT/customers/{id} - create", func(t *testing.T) {
		resetDatabase(t)
		resp, _ := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", customerURI, sampleCustomer.ID), sampleCustomer)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})
	t.Run("PATCH/customers/{id}", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)

		patch := map[string]any{"name": "PatchedName"}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", customerURI, newCustomer.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", customerURI, newCustomer.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var customer models.Customer
		err := json.Unmarshal(body, &customer)
		require.NoError(t, err)
		require.Equal(t, "PatchedName", customer.Name)
	})
	t.Run("DELETE/customers/{id}", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", customerURI, newCustomer.ID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", customerURI, newCustomer.ID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestRoomEndpoints(t *testing.T) {
	t.Run("POST/rooms", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)
		require.Equal(t, sampleRoom.Number, newRoom.Number)
	})
	t.Run("GET/rooms", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)

		resp, body := makeRequest(t, http.MethodGet, roomURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var rooms []models.Room
		err := json.Unmarshal(body, &rooms)
		require.NoError(t, err)
		require.Contains(t, rooms, newRoom)
	})
	t.Run("GET/rooms/{id}", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var room models.Room
		err := json.Unmarshal(body, &room)
		require.NoError(t, err)
		require.Equal(t, newRoom, room)
	})
	t.Run("PUT/rooms/{id} - update", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)
		newRoom.Price += 10

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), newRoom)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var room models.Room
		err := json.Unmarshal(body, &room)
		require.NoError(t, err)
		require.Equal(t, newRoom, room)
	})
	t.Run("PUT/rooms/{id} - create", func(t *testing.T) {
		resetDatabase(t)
		resp, _ := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", roomURI, sampleRoom.ID), sampleRoom)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})
	t.Run("PATCH/rooms/{id}", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)

		patch := map[string]any{"price": newRoom.Price + 20}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var room models.Room
		err := json.Unmarshal(body, &room)
		require.NoError(t, err)
		require.Equal(t, newRoom.Price+20, room.Price)
	})
	t.Run("DELETE/rooms/{id}", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestBookingEndpoints(t *testing.T) {
	setupDependencies := func(t *testing.T) models.BookingDTO {
		resetDatabase(t)
		booking := sampleBookingDTO
		booking.CustomerID = createSample(t, customerURI, sampleCustomer).ID
		booking.RoomID = createSample(t, roomURI, sampleRoom).ID
		return booking
	}
	t.Run("POST/bookings - success", func(t *testing.T) {
		booking := setupDependencies(t)

		newBooking := createSample(t, bookingURI, booking)
		require.Equal(t, booking.Code, newBooking.Code)
	})
	// test for validation logic
	t.Run("POST/bookings - start date must be before end date", func(t *testing.T) {
		invalidBooking := setupDependencies(t)

		invalidBooking.StartDate = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
		resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/bookings - start date must be in the future", func(t *testing.T) {
		invalidBooking := setupDependencies(t)

		invalidBooking.StartDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
		resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/bookings - start date and end date cannot be the same", func(t *testing.T) {
		invalidBooking := setupDependencies(t)

		invalidBooking.StartDate = invalidBooking.EndDate
		resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/bookings - customer does not exist", func(t *testing.T) {
		invalidBooking := setupDependencies(t)

		invalidBooking.CustomerID = -1
		resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/bookings - room does not exist", func(t *testing.T) {
		invalidBooking := setupDependencies(t)

		invalidBooking.RoomID = -1
		resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/bookings - booking code already exists", func(t *testing.T) {
		invalidBooking := setupDependencies(t)
		invalidBooking.Code = "UNIQUE123"
		resp, _ := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		invalidBooking.StartDate = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
		invalidBooking.EndDate = time.Now().AddDate(0, 1, 1).Format("2006-01-02")
		resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/bookings - booking dates overlap with an existing booking for the same room", func(t *testing.T) {
		invalidBooking := setupDependencies(t)
		createSample(t, bookingURI, invalidBooking)

		invalidBooking.ID = -1 // ensure it's treated as a new booking
		invalidBooking.Code = "OVERLAP123"
		invalidBooking.StartDate = time.Now().AddDate(0, 0, 3).Format("2006-01-02")
		resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("GET/bookings", func(t *testing.T) {
		booking := setupDependencies(t)
		booking = createSample(t, bookingURI, booking)

		resp, body := makeRequest(t, http.MethodGet, bookingURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var bookings []models.BookingDTO
		err := json.Unmarshal(body, &bookings)
		require.NoError(t, err)
		require.Contains(t, bookings, booking)
	})
	t.Run("GET/bookings/{id}", func(t *testing.T) {
		booking := setupDependencies(t)
		booking = createSample(t, bookingURI, booking)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", bookingURI, booking.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var b models.BookingDTO
		err := json.Unmarshal(body, &b)
		require.NoError(t, err)
		require.Equal(t, booking, b)
	})
	t.Run("PUT/bookings/{id} - update", func(t *testing.T) {
		booking := setupDependencies(t)
		booking = createSample(t, bookingURI, booking)
		booking.StartDate = time.Now().AddDate(0, 0, 3).Format("2006-01-02") // this also tests the overlapping logic on update

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", bookingURI, booking.ID), booking)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var b models.BookingDTO
		err := json.Unmarshal(body, &b)
		require.NoError(t, err)
		require.Equal(t, booking, b)
	})
	t.Run("PUT/bookings/{id} - create", func(t *testing.T) {
		booking := setupDependencies(t)
		resp, _ := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", bookingURI, booking.ID), booking)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})
	t.Run("PATCH/bookings/{id}", func(t *testing.T) {
		booking := setupDependencies(t)
		booking = createSample(t, bookingURI, booking)

		patch := map[string]any{"code": "PatchedCode123"}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", bookingURI, booking.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", bookingURI, booking.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var b models.BookingDTO
		err := json.Unmarshal(body, &b)
		require.NoError(t, err)
		require.Equal(t, "PatchedCode123", b.Code)
	})
	t.Run("DELETE/bookings/{id}", func(t *testing.T) {
		booking := setupDependencies(t)
		booking = createSample(t, bookingURI, booking)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", bookingURI, booking.ID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", bookingURI, booking.ID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestReviewEndpoints(t *testing.T) {
	setupDependencies := func(t *testing.T) models.ReviewDTO {
		resetDatabase(t)
		review := sampleReviewDTO
		booking := sampleBookingDTO
		booking.CustomerID = createSample(t, customerURI, sampleCustomer).ID
		booking.RoomID = createSample(t, roomURI, sampleRoom).ID
		review.BookingID = createSample(t, bookingURI, booking).ID
		return review
	}
	t.Run("POST/reviews - success", func(t *testing.T) {
		review := setupDependencies(t)
		newReview := createSample(t, reviewURI, review)
		require.Equal(t, review, newReview)
	})
	// test for validation logic
	t.Run("POST/reviews - booking does not exist", func(t *testing.T) {
		resetDatabase(t)
		invalidReview := sampleReviewDTO
		invalidReview.BookingID = -1
		resp, body := makeRequest(t, http.MethodPost, reviewURI, invalidReview)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/reviews - review date must be after booking start date", func(t *testing.T) {
		resetDatabase(t)
		booking := sampleBookingDTO
		newCustomer := createSample(t, customerURI, sampleCustomer)
		booking.CustomerID = newCustomer.ID
		newRoom := createSample(t, roomURI, sampleRoom)
		booking.RoomID = newRoom.ID
		newBookingDTO := createSample(t, bookingURI, booking)

		bdate, err := time.Parse("2006-01-02", newBookingDTO.StartDate)
		require.NoError(t, err)

		invalidReview := sampleReviewDTO
		invalidReview.Date = bdate.AddDate(0, 0, -1).Format("2006-01-02")

		invalidReview.BookingID = newBookingDTO.ID
		resp, body := makeRequest(t, http.MethodPost, reviewURI, invalidReview)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("POST/reviews - customer has already written a review for this booking", func(t *testing.T) {
		review := setupDependencies(t)
		review = createSample(t, reviewURI, review)

		resp, body := makeRequest(t, http.MethodPost, reviewURI, review)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("GET/reviews", func(t *testing.T) {
		review := setupDependencies(t)
		review = createSample(t, reviewURI, review)

		resp, body := makeRequest(t, http.MethodGet, reviewURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var reviews []models.ReviewDTO
		err := json.Unmarshal(body, &reviews)
		require.NoError(t, err)
		require.Contains(t, reviews, review)
	})
	t.Run("GET/reviews/{id}", func(t *testing.T) {
		review := setupDependencies(t)
		review = createSample(t, reviewURI, review)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", reviewURI, review.BookingID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var r models.ReviewDTO
		err := json.Unmarshal(body, &r)
		require.NoError(t, err)
		require.Equal(t, review, r)
	})
	t.Run("PUT/reviews/{id} - update", func(t *testing.T) {
		review := setupDependencies(t)
		review = createSample(t, reviewURI, review)
		review.Comment = "UpdatedComment"

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", reviewURI, review.BookingID), review) // this also test the validation logic on update (booking already reviewed)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var r models.ReviewDTO
		err := json.Unmarshal(body, &r)
		require.NoError(t, err)
		require.Equal(t, review, r)
	})
	t.Run("PUT/reviews/{id} - create", func(t *testing.T) {
		review := setupDependencies(t)
		resp, _ := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", reviewURI, review.BookingID), review)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})
	t.Run("PATCH/reviews/{id}", func(t *testing.T) {
		review := setupDependencies(t)
		review = createSample(t, reviewURI, review)

		patch := map[string]any{"comment": "PatchedComment"}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", reviewURI, review.BookingID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", reviewURI, review.BookingID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var r models.ReviewDTO
		err := json.Unmarshal(body, &r)
		require.NoError(t, err)
		require.Equal(t, "PatchedComment", r.Comment)
	})
	t.Run("DELETE/reviews/{id}", func(t *testing.T) {
		review := setupDependencies(t)
		review = createSample(t, reviewURI, review)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", reviewURI, review.BookingID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", reviewURI, review.BookingID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestHotelServiceEndpoints(t *testing.T) {
	t.Run("POST/services", func(t *testing.T) {
		resetDatabase(t)
		newService := createSample(t, serviceURI, sampleService)
		require.Equal(t, sampleService.Type, newService.Type)
	})
	t.Run("POST/services - service type already exists", func(t *testing.T) {
		resetDatabase(t)
		newService := createSample(t, serviceURI, sampleService)
		require.Equal(t, sampleService.Type, newService.Type)

		resp, body := makeRequest(t, http.MethodPost, serviceURI, sampleService)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		fmt.Print(string(body))
	})
	t.Run("GET/services", func(t *testing.T) {
		resetDatabase(t)
		newService := createSample(t, serviceURI, sampleService)

		resp, body := makeRequest(t, http.MethodGet, serviceURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var services []models.HotelService
		err := json.Unmarshal(body, &services)
		require.NoError(t, err)
		require.Contains(t, services, newService)
	})
	t.Run("GET/services/{id}", func(t *testing.T) {
		resetDatabase(t)
		newService := createSample(t, serviceURI, sampleService)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", serviceURI, newService.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var service models.HotelService
		err := json.Unmarshal(body, &service)
		require.NoError(t, err)
		require.Equal(t, newService, service)
	})
	t.Run("PUT/services/{id} - update", func(t *testing.T) {
		resetDatabase(t)
		newService := createSample(t, serviceURI, sampleService)
		newService.Description = "UpdatedDescription"

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", serviceURI, newService.ID), newService)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var service models.HotelService
		err := json.Unmarshal(body, &service)
		require.NoError(t, err)
		require.Equal(t, newService, service)
	})
	t.Run("PUT/services/{id} - create", func(t *testing.T) {
		resetDatabase(t)
		resp, _ := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", serviceURI, sampleService.ID), sampleService)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})
	t.Run("PATCH/services/{id}", func(t *testing.T) {
		resetDatabase(t)
		newService := createSample(t, serviceURI, sampleService)

		patch := map[string]any{"description": "PatchedDescription"}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", serviceURI, newService.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", serviceURI, newService.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var service models.HotelService
		err := json.Unmarshal(body, &service)
		require.NoError(t, err)
		require.Equal(t, "PatchedDescription", service.Description)
	})
	t.Run("DELETE/services/{id}", func(t *testing.T) {
		resetDatabase(t)
		newService := createSample(t, serviceURI, sampleService)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", serviceURI, newService.ID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", serviceURI, newService.ID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestServiceRequestEndpoints(t *testing.T) {
	setupDependencies := func(t *testing.T) models.ServiceRequestDTO {
		resetDatabase(t)
		booking := sampleBookingDTO
		request := sampleServiceRequestDTO
		booking.CustomerID = createSample(t, customerURI, sampleCustomer).ID
		booking.RoomID = createSample(t, roomURI, sampleRoom).ID
		createSample(t, bookingURI, booking)
		request.CustomerID = booking.CustomerID
		request.ServiceID = createSample(t, serviceURI, sampleService).ID
		return request
	}
	t.Run("POST/service-requests - success", func(t *testing.T) {
		request := setupDependencies(t)
		newRequest := createSample(t, requestURI, request)
		require.Equal(t, request.Date, newRequest.Date)
	})
	// test for validation logic
	t.Run("POST/service-requests - service request date must be in the future", func(t *testing.T) {
		request := setupDependencies(t)
		request.Date = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		resp, body := makeRequest(t, http.MethodPost, requestURI, request)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		fmt.Print(string(body))
	})
	t.Run("POST/service-requests - customer does not exist", func(t *testing.T) {
		request := setupDependencies(t)
		request.CustomerID = -1
		resp, body := makeRequest(t, http.MethodPost, requestURI, request)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		fmt.Print(string(body))
	})
	t.Run("POST/service-requests - customer has no bookings", func(t *testing.T) {
		resetDatabase(t)
		request := sampleServiceRequestDTO
		request.CustomerID = createSample(t, customerURI, sampleCustomer).ID
		request.ServiceID = createSample(t, serviceURI, sampleService).ID
		resp, body := makeRequest(t, http.MethodPost, requestURI, request)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		fmt.Print(string(body))
	})
	t.Run("POST/service-requests - service request date must be within a booking period", func(t *testing.T) {
		request := setupDependencies(t)
		request.Date = time.Now().AddDate(0, 0, 10).Format("2006-01-02")
		resp, body := makeRequest(t, http.MethodPost, requestURI, request)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		fmt.Print(string(body))
	})
	t.Run("POST/service-requests - service does not exist", func(t *testing.T) {
		request := setupDependencies(t)
		request.ServiceID = -1
		resp, body := makeRequest(t, http.MethodPost, requestURI, request)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		fmt.Print(string(body))
	})
	t.Run("POST/service-requests - duplicate service request", func(t *testing.T) {
		request := setupDependencies(t)
		createSample(t, requestURI, request)
		resp, body := makeRequest(t, http.MethodPost, requestURI, request)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		fmt.Print(string(body))
	})
	t.Run("GET/service-requests", func(t *testing.T) {
		request := setupDependencies(t)
		request = createSample(t, requestURI, request)

		resp, body := makeRequest(t, http.MethodGet, requestURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var requests []models.ServiceRequestDTO
		err := json.Unmarshal(body, &requests)
		require.NoError(t, err)
		require.Contains(t, requests, request)
	})
	t.Run("GET/service-requests/{id}", func(t *testing.T) {
		request := setupDependencies(t)
		request = createSample(t, requestURI, request)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", requestURI, request.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var r models.ServiceRequestDTO
		err := json.Unmarshal(body, &r)
		require.NoError(t, err)
		require.Equal(t, request, r)
	})
	t.Run("PUT/service-requests/{id} - update", func(t *testing.T) {
		request := setupDependencies(t)
		request = createSample(t, requestURI, request)
		request.Date = time.Now().AddDate(0, 0, 4).Format("2006-01-02")

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", requestURI, request.ID), request)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var r models.ServiceRequestDTO
		err := json.Unmarshal(body, &r)
		require.NoError(t, err)
		require.Equal(t, request, r)
	})
	t.Run("PUT/service-requests/{id} - create", func(t *testing.T) {
		request := setupDependencies(t)
		resp, _ := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", requestURI, request.ID), request)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})
	t.Run("PATCH/service-requests/{id}", func(t *testing.T) {
		request := setupDependencies(t)
		request = createSample(t, requestURI, request)

		patch := map[string]any{"date": time.Now().AddDate(0, 0, 4).Format("2006-01-02")}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", requestURI, request.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", requestURI, request.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var r models.ServiceRequestDTO
		err := json.Unmarshal(body, &r)
		require.NoError(t, err)
		require.Equal(t, time.Now().AddDate(0, 0, 4).Format("2006-01-02"), r.Date)
	})
	t.Run("DELETE/service-requests/{id}", func(t *testing.T) {
		request := setupDependencies(t)
		request = createSample(t, requestURI, request)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", requestURI, request.ID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", requestURI, request.ID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// questo può funzionare se fai un'interfaccia comune per tutti i modelli per fare il get e il set dell'ID, non lo faccio perché non cambio la logica del server per i test, anche se potrebbe migliorare
// func TestAllEntityCRUD(t *testing.T) {
// 	testCases := []struct {
// 		name       string
// 		endpoint   string
// 		sampleData any
// 		updateData any
// 	}{
// 		{
// 			name:     "customers",
// 			endpoint: "/customers",
// 			sampleData: models.Customer{
// 				CF:    "TESTCF12345",
// 				Name:  "Testino",
// 				Age:   30,
// 				Email: "testcustomer@example.com",
// 			},
// 			updateData: models.Customer{
// 				CF:    "TESTCF12345",
// 				Name:  "UpdatedName",
// 				Age:   30,
// 				Email: "testcustomer@example.com",
// 			},
// 		},
// 		{
// 			name:     "bookings",
// 			endpoint: "/bookings",
// 			sampleData: models.BookingDTO{
// 				Code:      "TESTBOOK123",
// 				customerID:  1,
// 				RoomID:    1,
// 				StartDate: "2026-10-01",
// 				EndDate:   "2026-10-05",
// 			},
// 			updateData: models.BookingDTO{
// 				Code:      "UpdatedCode",
// 				customerID:  1,
// 				RoomID:    1,
// 				StartDate: "2026-10-01",
// 				EndDate:   "2026-10-05",
// 			},
// 		},
// 	}
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			uri := baseURI + tc.endpoint
// 			t.Run("POST"+tc.endpoint, func(t *testing.T) {
// 				resetDatabase(t)
// 				newSample := createSample(t, uri, tc.sampleData)
// 				tc.sampleData.ID = newSample.ID
// 				require.Equal(t, tc.sampleData, newSample)
// 			})
// 			t.Run("GET"+tc.endpoint, func(t *testing.T) {
// 				resetDatabase(t)
// 				newSample := createSample(t, uri, tc.sampleData)
// 				tc.sampleData.ID = newSample.ID
// 				resp, body := makeRequest(t, http.MethodGet, uri, nil)
// 				require.Equal(t, http.StatusOK, resp.StatusCode)
// 				var samples any
// 				err := json.Unmarshal(body, &samples)
// 				require.NoError(t, err)
// 				require.Contains(t, samples, newSample)
// 			})
// 			t.Run("GET"+tc.endpoint+"/{id}", func(t *testing.T) {
// 				resetDatabase(t)
// 				newSample := createSample(t, uri, tc.sampleData)
// 				tc.sampleData.ID = newSample.ID
// 				resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", uri, tc.sampleData.ID), nil)
// 				require.Equal(t, http.StatusOK, resp.StatusCode)
// 				var sample any
// 				err := json.Unmarshal(body, &sample)
// 				require.NoError(t, err)
// 				require.Equal(t, tc.sampleData, sample)
// 			})
// 			t.Run("PUT"+tc.endpoint+"/{id}", func(t *testing.T) {
// 				resetDatabase(t)
// 				newSample := createSample(t, uri, tc.sampleData)
// 				tc.sampleData.ID = newSample.ID
// 				resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", uri, tc.sampleData.ID), tc.updateData)
// 				require.Equal(t, http.StatusOK, resp.StatusCode)
// 				var sample any
// 				err := json.Unmarshal(body, &sample)
// 				require.NoError(t, err)
// 				require.Equal(t, tc.updateData, sample)
// 			})
// 			t.Run("DELETE"+tc.endpoint+"/{id}", func(t *testing.T) {
// 				resetDatabase(t)
// 				newSample := createSample(t, uri, tc.sampleData)
// 				tc.sampleData.ID = newSample.ID
// 				resp, body := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", uri, tc.sampleData.ID), nil)
// 				require.Equal(t, http.StatusOK, resp.StatusCode)
// 				var sample any
// 				err := json.Unmarshal(body, &sample)
// 				require.NoError(t, err)
// 				require.Equal(t, tc.sampleData, sample)
// 				resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", uri, tc.sampleData.ID), nil)
// 				require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 			})
// 		})
// 	}
// }
