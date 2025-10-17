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
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

var (
	conn         *pgx.Conn
	baseURI      string
	populatePath string
	client       = &http.Client{}
)

func getLocalDBConnStr(admin bool) string {
	host := os.Getenv("TEST_DB_HOST")
	port := os.Getenv("TEST_DB_PORT")
	user := os.Getenv("TEST_DB_USER")
	password := os.Getenv("TEST_DB_PASSWORD")
	dbname := os.Getenv("TEST_DB_NAME")

	if admin {
		dbname = os.Getenv("DB_NAME")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func TestMain(m *testing.M) {
	// connect to the admin database and create a new test database
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		os.Exit(1)
	}
	populatePath = os.Getenv("SQL_POPULATE_PATH")
	ctx := context.Background()
	adminConn, err := pgx.Connect(ctx, getLocalDBConnStr(true))
	if err != nil {
		fmt.Println("Unable to connect to admin database:", err)
		os.Exit(1)
	}
	defer adminConn.Close(ctx)

	_, err = adminConn.Exec(ctx, "DROP DATABASE IF EXISTS testdb")
	if err != nil {
		fmt.Println("Unable to drop test database:", err)
		os.Exit(1)
	}
	_, err = adminConn.Exec(ctx, "CREATE DATABASE testdb TEMPLATE temp_postgres")
	if err != nil {
		fmt.Println("Unable to create test database:", err)
		os.Exit(1)
	}
	conn, err = pgx.Connect(ctx, getLocalDBConnStr(false))
	if err != nil {
		fmt.Println("Unable to connect to test database:", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	val := validator.New()
	mux := http.NewServeMux()
	setupRoutes(mux, conn, val)
	testServer := httptest.NewServer(mux)
	baseURI = testServer.URL
	defer testServer.Close()

	code := m.Run()

	os.Exit(code)
}

func TestHelloWorld(t *testing.T) {
	resp, err := http.Get(baseURI + "/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// truncate all tables and repopulate with initial data from populate.sql
func resetDatabase(t *testing.T) {
	ctx := context.Background()
	_, err := conn.Exec(ctx, "TRUNCATE TABLE customer, booking, review, service_request, hotel_service, room RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to truncate tables: %v", err)
	sql, err := os.ReadFile(populatePath)
	require.NoError(t, err, "Failed to read %s: %v", populatePath, err)
	_, err = conn.Exec(ctx, string(sql))
	require.NoError(t, err, "Failed to execute %s: %v", populatePath, err)
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
	customerURI := baseURI + "/customers"
	sampleCustomer := models.Customer{
		CF:    "TESTCF12345",
		Name:  "Testino",
		Age:   30,
		Email: "testcustomer@example.com",
	}
	t.Run("POST/customers", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)
		sampleCustomer.ID = newCustomer.ID
		require.Equal(t, sampleCustomer, newCustomer)
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
	t.Run("PUT/customers/{id}", func(t *testing.T) {
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
	t.Run("PATCH/customers/{id}", func(t *testing.T) {
		resetDatabase(t)
		newCustomer := createSample(t, customerURI, sampleCustomer)

		patch := map[string]any{"name": "PatchedName"}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", customerURI, newCustomer.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
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

func TestBookingEndpoints(t *testing.T) {
	bookingURI := baseURI + "/bookings"
	sampleBookingDTO := models.BookingDTO{
		Code:       "TESTBOOK123",
		CustomerID: 1,
		RoomID:     1,
		StartDate:  time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		EndDate:    time.Now().AddDate(0, 0, 8).Format("2006-01-02"),
	}
	t.Run("POST/bookings", func(t *testing.T) {
		resetDatabase(t)
		newBooking := createSample(t, bookingURI, sampleBookingDTO)
		sampleBookingDTO.ID = newBooking.ID
		require.Equal(t, sampleBookingDTO, newBooking)

		// test for validation logic
		t.Run("start date > end date", func(t *testing.T) {
			resetDatabase(t)
			invalidBooking := sampleBookingDTO
			invalidBooking.StartDate = time.Now().AddDate(0, 1, 0).Format("2006-01-02")
			resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		})
		t.Run("start date < current date", func(t *testing.T) {
			resetDatabase(t)
			invalidBooking := sampleBookingDTO
			invalidBooking.StartDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
			resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		})
		t.Run("start date = end date", func(t *testing.T) {
			resetDatabase(t)
			invalidBooking := sampleBookingDTO
			invalidBooking.StartDate = invalidBooking.EndDate
			resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		})
		t.Run("customer does not exist", func(t *testing.T) {
			resetDatabase(t)
			invalidBooking := sampleBookingDTO
			invalidBooking.CustomerID = -1
			resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		})
		t.Run("room does not exist", func(t *testing.T) {
			resetDatabase(t)
			invalidBooking := sampleBookingDTO
			invalidBooking.RoomID = -1
			resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		})
		t.Run("overlapping bookings", func(t *testing.T) {
			resetDatabase(t)
			createSample(t, bookingURI, sampleBookingDTO)
			invalidBooking := sampleBookingDTO
			invalidBooking.ID = -1 // ensure it's treated as a new booking
			invalidBooking.Code = "OVERLAP123"
			invalidBooking.StartDate = time.Now().AddDate(0, 0, 3).Format("2006-01-02")
			resp, body := makeRequest(t, http.MethodPost, bookingURI, invalidBooking)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error, got: %s", string(body))
		})
	})
	t.Run("GET/bookings", func(t *testing.T) {
		resetDatabase(t)
		newBooking := createSample(t, bookingURI, sampleBookingDTO)

		resp, body := makeRequest(t, http.MethodGet, bookingURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var bookings []models.BookingDTO
		err := json.Unmarshal(body, &bookings)
		require.NoError(t, err)
		require.Contains(t, bookings, newBooking)
	})
	t.Run("GET/bookings/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBooking := createSample(t, bookingURI, sampleBookingDTO)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", bookingURI, newBooking.ID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var booking models.BookingDTO
		err := json.Unmarshal(body, &booking)
		require.NoError(t, err)
		require.Equal(t, newBooking, booking)
	})
	t.Run("PUT/bookings/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBooking := createSample(t, bookingURI, sampleBookingDTO)
		newBooking.StartDate = time.Now().AddDate(0, 0, 3).Format("2006-01-02") // this also tests the overlapping logic on update

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", bookingURI, newBooking.ID), newBooking)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var booking models.BookingDTO
		err := json.Unmarshal(body, &booking)
		require.NoError(t, err)
		require.Equal(t, newBooking, booking)
	})
	t.Run("PATCH/bookings/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBooking := createSample(t, bookingURI, sampleBookingDTO)

		patch := map[string]any{"code": "PatchedCode123"}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", bookingURI, newBooking.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("DELETE/bookings/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBooking := createSample(t, bookingURI, sampleBookingDTO)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", bookingURI, newBooking.ID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", bookingURI, newBooking.ID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestReviewEndpoints(t *testing.T) {
	bookingURI := baseURI + "/bookings"
	sampleBookingDTO := models.BookingDTO{
		Code:       "TESTBOOK123",
		CustomerID: 1,
		RoomID:     1,
		StartDate:  time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		EndDate:    time.Now().AddDate(0, 0, 8).Format("2006-01-02"),
	}
	reviewURI := baseURI + "/reviews"
	sampleReviewDTO := models.ReviewDTO{
		BookingID: -1,
		Comment:   "comment",
		Rating:    3,
		Date:      time.Now().AddDate(0, 0, 9).Format("2006-01-02"),
	}
	t.Run("POST/reviews", func(t *testing.T) {
		resetDatabase(t)
		newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

		sampleReviewDTO.BookingID = newBookingDTO.ID
		newReview := createSample(t, reviewURI, sampleReviewDTO)
		require.Equal(t, sampleReviewDTO, newReview)

		// test for validation logic
		t.Run("booking does not exist", func(t *testing.T) {
			resetDatabase(t)
			invalidReview := sampleReviewDTO
			invalidReview.BookingID = -1
			resp, _ := makeRequest(t, http.MethodPost, reviewURI, invalidReview)
			require.Equal(t, resp.StatusCode, http.StatusBadRequest)
		})
		t.Run("review date < booking start date", func(t *testing.T) {
			resetDatabase(t)
			newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

			bdate, err := time.Parse("2006-01-02", newBookingDTO.StartDate)
			require.NoError(t, err)

			invalidReview := sampleReviewDTO
			invalidReview.Date = bdate.AddDate(0, 0, -1).Format("2006-01-02")

			resp, _ := makeRequest(t, http.MethodPost, reviewURI, invalidReview)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
		t.Run("customer already reviewed", func(t *testing.T) {
			resetDatabase(t)
			newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

			sampleReviewDTO.BookingID = newBookingDTO.ID
			newReview := createSample(t, reviewURI, sampleReviewDTO)

			invalidReview := newReview
			resp, _ := makeRequest(t, http.MethodPost, reviewURI, invalidReview)
			require.Equal(t, resp.StatusCode, http.StatusBadRequest)
		})
	})
	t.Run("GET/reviews", func(t *testing.T) {
		resetDatabase(t)
		newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

		sampleReviewDTO.BookingID = newBookingDTO.ID
		newReview := createSample(t, reviewURI, sampleReviewDTO)

		resp, body := makeRequest(t, http.MethodGet, reviewURI, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var reviews []models.ReviewDTO
		err := json.Unmarshal(body, &reviews)
		require.NoError(t, err)
		require.Contains(t, reviews, newReview)
	})
	t.Run("GET/reviews/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

		sampleReviewDTO.BookingID = newBookingDTO.ID
		newReview := createSample(t, reviewURI, sampleReviewDTO)

		resp, body := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", reviewURI, newReview.BookingID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var review models.ReviewDTO
		err := json.Unmarshal(body, &review)
		require.NoError(t, err)
		require.Equal(t, newReview, review)
	})
	t.Run("PUT/reviews/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

		sampleReviewDTO.BookingID = newBookingDTO.ID // this also test the validation logic on update (booking already reviewed)
		newReview := createSample(t, reviewURI, sampleReviewDTO)
		newReview.Comment = "UpdatedComment"

		resp, body := makeRequest(t, http.MethodPut, fmt.Sprintf("%s/%d", reviewURI, newReview.BookingID), newReview)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var review models.ReviewDTO
		err := json.Unmarshal(body, &review)
		require.NoError(t, err)
		require.Equal(t, newReview, review)
	})
	t.Run("PATCH/reviews/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

		sampleReviewDTO.BookingID = newBookingDTO.ID
		newReview := createSample(t, reviewURI, sampleReviewDTO)

		patch := map[string]any{"comment": "PatchedComment"}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", reviewURI, newReview.BookingID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("DELETE/reviews/{id}", func(t *testing.T) {
		resetDatabase(t)
		newBookingDTO := createSample(t, bookingURI, sampleBookingDTO)

		sampleReviewDTO.BookingID = newBookingDTO.ID
		newReview := createSample(t, reviewURI, sampleReviewDTO)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", reviewURI, newReview.BookingID), nil)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", reviewURI, newReview.BookingID), nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestRoomEndpoints(t *testing.T) {
	roomURI := baseURI + "/rooms"
	sampleRoom := models.Room{
		Number:   101,
		Type:     "basic",
		Price:    100,
		Capacity: 2,
	}
	t.Run("POST/rooms", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)
		sampleRoom.ID = newRoom.ID
		require.Equal(t, sampleRoom, newRoom)
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
	t.Run("PUT/rooms/{id}", func(t *testing.T) {
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
	t.Run("PATCH/rooms/{id}", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)

		patch := map[string]any{"price": newRoom.Price + 20}
		resp, _ := makeRequest(t, http.MethodPatch, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), patch)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
	t.Run("DELETE/rooms/{id}", func(t *testing.T) {
		resetDatabase(t)
		newRoom := createSample(t, roomURI, sampleRoom)

		resp, _ := makeRequest(t, http.MethodDelete, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), newRoom)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		resp, _ = makeRequest(t, http.MethodGet, fmt.Sprintf("%s/%d", roomURI, newRoom.ID), nil)
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
