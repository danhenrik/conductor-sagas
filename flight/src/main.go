package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/conductor-sdk/conductor-go/sdk/client"
	"github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/conductor-sdk/conductor-go/sdk/settings"
	"github.com/conductor-sdk/conductor-go/sdk/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ####################################################################################################################################
// ### Models 																																																										###
// ####################################################################################################################################

type BookingStatus string

const (
	BookingStatusActive   BookingStatus = "active"
	BookingStatusCanceled BookingStatus = "canceled"
)

type Flight struct {
	FlightID       string    `json:"flightId" db:"flight_id"`
	Airline        string    `json:"airline" db:"airline"`
	Origin         string    `json:"origin" db:"origin"`
	Destination    string    `json:"destination" db:"destination"`
	DepartureTime  time.Time `json:"departureTime" db:"departure_time"`
	ArrivalTime    time.Time `json:"arrivalTime" db:"arrival_time"`
	Capacity       int       `json:"capacity" db:"capacity"`
	AvailableSeats int       `json:"availableSeats" db:"available_seats"`
}

type FlightBooking struct {
	BookingID     string        `json:"bookingId" db:"booking_id"`
	FlightID      string        `json:"flightId" db:"flight_id"`
	CustomerName  string        `json:"customerName" db:"customer_name"`
	CustomerEmail string        `json:"customerEmail" db:"customer_email"`
	SeatNumber    int           `json:"seatNumber" db:"seat_number"`
	BookingStatus BookingStatus `json:"bookingStatus" db:"booking_status"`
	BookingTime   time.Time     `json:"bookingTime" db:"booking_time"`
	UpdatedAt     time.Time     `json:"updateAt" db:"updated_at"`
}

// ####################################################################################################################################
// ### Database 																																																										###
// ####################################################################################################################################

var db *pgxpool.Pool

func connectDB() {
	databaseURL := "host=localhost user=saga-flight password=saga-flight dbname=saga-flight sslmode=disable" // e.g., "postgres://user:password@localhost:5432/dbname"
	// databaseURL := "host=database-flight user=saga-flight password=saga-flight dbname=saga-flight sslmode=disable" // e.g., "postgres://user:password@localhost:5432/dbname"
	var err error

	db, err = pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Connected to PostgreSQL!")
}

func closeDB() {
	db.Close()
	log.Println("Database connection closed.")
}

// ####################################################################################################################################
// ### Request Objects 																																																						###
// ####################################################################################################################################

type CreateFlightRequest struct {
	Airline       string    `json:"airline" binding:"required"`
	Origin        string    `json:"origin" binding:"required"`
	Destination   string    `json:"destination" binding:"required"`
	DepartureTime time.Time `json:"departureTime" binding:"required"`
	ArrivalTime   time.Time `json:"arrivalTime" binding:"required"`
	Capacity      int       `json:"capacity" binding:"required"`
}

type CreateBookingRequest struct {
	FlightID      string `json:"flightId" binding:"required"`
	CustomerName  string `json:"customerName" binding:"required"`
	CustomerEmail string `json:"customerEmail" binding:"required"`
	SeatNumber    int    `json:"seatNumber" binding:"required"`
}

// ####################################################################################################################################
// ### Handlers 																																																										###
// ####################################################################################################################################

// Flight
func createFlight(c *gin.Context) {
	var req CreateFlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newFlight := Flight{
		FlightID:       uuid.New().String(),
		Airline:        req.Airline,
		Origin:         req.Origin,
		Destination:    req.Destination,
		DepartureTime:  req.DepartureTime,
		ArrivalTime:    req.ArrivalTime,
		Capacity:       req.Capacity,
		AvailableSeats: req.Capacity,
	}

	_, err := db.Exec(context.Background(),
		"INSERT INTO flights (flight_id, airline, origin, destination, departure_time, arrival_time, capacity, available_seats) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		newFlight.FlightID, newFlight.Airline, newFlight.Origin, newFlight.Destination, newFlight.DepartureTime, newFlight.ArrivalTime, newFlight.Capacity, newFlight.AvailableSeats)

	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create flight"})
		return
	}

	c.JSON(http.StatusCreated, newFlight)
}

func getFlights(c *gin.Context) {
	rows, err := db.Query(context.Background(), "SELECT * FROM flights")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch flights"})
		return
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var flight Flight
		if err := rows.Scan(&flight.FlightID, &flight.Airline, &flight.Origin, &flight.Destination,
			&flight.DepartureTime, &flight.ArrivalTime, &flight.Capacity, &flight.AvailableSeats); err != nil {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning flight data"})
			return
		}
		flights = append(flights, flight)
	}

	c.JSON(http.StatusOK, flights)
}

func getFlightByID(c *gin.Context) {
	flightID := c.Param("id")

	var flight Flight
	err := db.QueryRow(context.Background(),
		"SELECT flight_id, airline, origin, destination, departure_time, arrival_time, capacity, available_seats FROM flights WHERE flight_id=$1", flightID).
		Scan(&flight.FlightID, &flight.Airline, &flight.Origin, &flight.Destination, &flight.DepartureTime, &flight.ArrivalTime, &flight.Capacity, &flight.AvailableSeats)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Flight not found"})
		} else {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch flight"})
		}
		return
	}

	c.JSON(http.StatusOK, flight)
}

func deleteFlight(c *gin.Context) {
	flightID := c.Param("id")

	_, err := db.Exec(context.Background(),
		"DELETE FROM flights WHERE flight_id=$1",
		flightID)
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete flight"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Flight deleted successfully"})
}

// FlightBooking
type TaskOutput struct {
	Data      map[string]string
	Success   bool
	ErrorCode int
}

func successOutput(data ...map[string]string) (*TaskOutput, error) {
	if len(data) != 1 {
		data = append(data, map[string]string{})
	}

	return &TaskOutput{
		Data:      data[0],
		Success:   true,
		ErrorCode: 0,
	}, nil
}

func errorOutput(err error, errorCode ...int) (*TaskOutput, error) {
	if len(errorCode) != 1 {
		errorCode = append(errorCode, 1)
	}

	return &TaskOutput{
		Data:      map[string]string{"error": err.Error()},
		Success:   false,
		ErrorCode: errorCode[0],
	}, nil
}

func createBookingWorker(task *model.Task) (interface{}, error) {
	customerEmail := task.InputData["customerEmail"].(string)
	customerName := task.InputData["customerName"].(string)
	flightID := task.InputData["flightId"].(string)
	seatNumber := int(task.InputData["seatNumber"].(float64))

	newBooking, err := createBooking(flightID, customerName, customerEmail, seatNumber)
	if err != nil {
		println(err.Error())
		return errorOutput(err)
	}

	body := map[string]string{
		"bookingId":     newBooking.BookingID,
		"bookingStatus": string(newBooking.BookingStatus),
		"bookingTime":   newBooking.BookingTime.String(),
		"updatedAt":     newBooking.UpdatedAt.String(),
	}

	return successOutput(body)
}

func createBookingHTTP(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBooking, err := createBooking(req.FlightID, req.CustomerName, req.CustomerEmail, req.SeatNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, newBooking)
}

func createBooking(flightId string, customerName string, customerEmail string, seatNumber int) (*FlightBooking, error) {
	var flightExists bool
	err := db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM flights WHERE flight_id=$1)", flightId).
		Scan(&flightExists)
	if err != nil {
		return nil, fmt.Errorf("Failed to validate flight")
	}

	if !flightExists {
		return nil, fmt.Errorf("Flight does not exist")
	}

	newBooking := FlightBooking{
		BookingID:     uuid.New().String(),
		FlightID:      flightId,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		SeatNumber:    seatNumber,
		BookingStatus: BookingStatusActive,
		BookingTime:   time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = db.Exec(context.Background(),
		"INSERT INTO flight_bookings (booking_id, flight_id, customer_name, customer_email, seat_number, booking_status, booking_time, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		newBooking.BookingID, newBooking.FlightID, newBooking.CustomerName, newBooking.CustomerEmail, newBooking.SeatNumber, newBooking.BookingStatus, newBooking.BookingTime, newBooking.UpdatedAt)

	if err != nil {
		if pgxErr, ok := err.(*pgconn.PgError); ok && pgxErr.Code == "23505" {
			return nil, fmt.Errorf("Seat already booked for this flight")
		}
		return nil, fmt.Errorf("Failed to create booking")
	}

	return &newBooking, nil
}

func getBookings(c *gin.Context) {
	bookingStatus := c.Query("bookingStatus")
	flightID := c.Query("flightId")
	bookingID := c.Query("bookingId")
	customerEmail := c.Query("customerEmail")

	query := "SELECT * FROM flight_bookings WHERE 1=1"
	args := []interface{}{}
	argID := 1

	if bookingStatus != "" {
		query += " AND booking_status=$" + strconv.Itoa(argID)
		args = append(args, bookingStatus)
		argID++
	}
	if flightID != "" {
		query += " AND flight_id=$" + strconv.Itoa(argID)
		args = append(args, flightID)
		argID++
	}
	if bookingID != "" {
		query += " AND booking_id=$" + strconv.Itoa(argID)
		args = append(args, bookingID)
		argID++
	}
	if customerEmail != "" {
		query += " AND customer_email=$" + strconv.Itoa(argID)
		args = append(args, customerEmail)
		argID++
	}

	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch bookings"})
		return
	}
	defer rows.Close()

	var bookings []FlightBooking
	for rows.Next() {
		var booking FlightBooking
		if err := rows.Scan(&booking.BookingID, &booking.FlightID, &booking.CustomerName, &booking.CustomerEmail,
			&booking.SeatNumber, &booking.BookingStatus, &booking.BookingTime, &booking.UpdatedAt); err != nil {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning booking data"})
			return
		}
		bookings = append(bookings, booking)
	}

	c.JSON(http.StatusOK, bookings)
}

// TODO: Get booking tão com problema
func getBookingByID(c *gin.Context) {
	bookingID := c.Param("id")

	var booking FlightBooking
	err := db.QueryRow(context.Background(),
		"SELECT booking_id, flight_id, customer_name, customer_email, seat_number, booking_status, booking_time, updated_at FROM flight_bookings WHERE booking_id=$1", bookingID).Scan(
		&booking.BookingID, &booking.FlightID, &booking.CustomerName, &booking.CustomerEmail, &booking.SeatNumber, &booking.BookingStatus, &booking.BookingTime, &booking.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		} else {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch booking"})
		}
		return
	}

	c.JSON(http.StatusOK, booking)
}

func deleteBookingByID(c *gin.Context) {
	bookingID := c.Param("id")

	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to cancel (Missing bookingId)"})
		return
	}

	println(bookingID)

	_, err := db.Exec(context.Background(),
		"UPDATE flight_bookings SET booking_status=$1, updated_at=$2 WHERE booking_id=$3",
		BookingStatusCanceled, time.Now(), bookingID)
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking canceled successfully"})
}

func deleteBookingByFlightSeatWorker(task *model.Task) (interface{}, error) {
	flightID := task.InputData["flightId"].(string)
	seatNumber := strconv.Itoa(int(task.InputData["seatNumber"].(float64)))

	err := deleteBookingByFlightSeat(flightID, seatNumber)
	if err != nil {
		return errorOutput(err)
	}

	return successOutput()
}

func deleteBookingByFlightSeatHTTP(c *gin.Context) {
	flightID := c.Param("flightId")
	seatNumber := c.Param("seatNumber")

	err := deleteBookingByFlightSeat(flightID, seatNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking canceled successfully"})
}

func deleteBookingByFlightSeat(flightId string, seatNumber string) error {
	if flightId == "" {
		return fmt.Errorf("Failed to cancel (Missing flightId)")
	}

	if seatNumber == "" {
		return fmt.Errorf("Failed to cancel (Missing seatNumber)")
	}

	_, err := db.Exec(context.Background(),
		"UPDATE flight_bookings SET booking_status=$1, updated_at=$2 WHERE flight_id=$3 AND seat_number=$4",
		BookingStatusCanceled, time.Now(), flightId, seatNumber)
	if err != nil {
		return err
	}

	return nil
}

// ####################################################################################################################################
// ### Main 																																																										###
// ####################################################################################################################################

func main() {
	connectDB()
	r := gin.Default()

	r.POST("/flights", createFlight)
	r.GET("/flights", getFlights)
	r.GET("/flights/:id", getFlightByID)
	r.DELETE("/flights/:id", deleteFlight)

	r.POST("/bookings", createBookingHTTP)
	r.GET("/bookings", getBookings)
	r.GET("/bookings/:id", getBookingByID)
	r.DELETE("/bookings/id/:id", deleteBookingByID)
	r.DELETE("/bookings/seat/:flightId/:seatNumber", deleteBookingByFlightSeatHTTP)

	// ############################################################
	// ### Conductor client setup
	// ############################################################
	var apiClient = client.NewAPIClient(
		nil,
		settings.NewHttpSettings("http://localhost:8080/api"),
	)
	var taskRunner = worker.NewTaskRunnerWithApiClient(apiClient)
	// WorkflowExecutor could be used to start workflows, possibly used in the separate implementation from compensation implementation
	// var workflowExecutor = executor.NewWorkflowExecutor(apiClient)

	// ### Workers
	taskRunner.StartWorker("book_flight", createBookingWorker, 1, time.Millisecond*100)
	taskRunner.StartWorker("cancel_flight_booking", deleteBookingByFlightSeatWorker, 1, time.Millisecond*100)

	r.Run(":3000")
}
