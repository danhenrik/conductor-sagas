package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

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
}

// ####################################################################################################################################
// ### Database 																																																										###
// ####################################################################################################################################

var db *pgxpool.Pool

func connectDB() {
	databaseURL := "host=localhost user=saga-flight password=saga-flight dbname=saga-flight sslmode=disable" // e.g., "postgres://user:password@localhost:5432/dbname"
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
// ### Request Objects 																																																						###
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
		log.Printf("Failed to execute query: %v", err)
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
	err := db.QueryRow(context.Background(), "SELECT flight_id, airline, origin, destination, departure_time, arrival_time, capacity, available_seats FROM flights WHERE flight_id=$1", flightID).Scan(
		&flight.FlightID, &flight.Airline, &flight.Origin, &flight.Destination, &flight.DepartureTime, &flight.ArrivalTime, &flight.Capacity, &flight.AvailableSeats)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Flight not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch flight"})
		}
		return
	}

	c.JSON(http.StatusOK, flight)
}

func deleteFlight(c *gin.Context) {
	flightID := c.Param("id")

	_, err := db.Exec(context.Background(), "DELETE FROM flights WHERE flight_id=$1", flightID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete flight"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Flight deleted successfully"})
}

// FlightBooking
func createBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBooking := FlightBooking{
		BookingID:     uuid.New().String(),
		FlightID:      req.FlightID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		SeatNumber:    req.SeatNumber,
		BookingStatus: BookingStatusActive,
		BookingTime:   time.Now(),
	}

	_, err := db.Exec(context.Background(),
		"INSERT INTO flight_bookings (booking_id, flight_id, customer_name, customer_email, seat_number, booking_status, booking_time) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		newBooking.BookingID, newBooking.FlightID, newBooking.CustomerName, newBooking.CustomerEmail, newBooking.SeatNumber, newBooking.BookingStatus, newBooking.BookingTime)

	if err != nil {
		if pgxErr, ok := err.(*pgconn.PgError); ok && pgxErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "Seat already booked for this flight"})
		} else {
			log.Printf("Failed to execute query: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		}
		return
	}

	c.JSON(http.StatusCreated, newBooking)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch bookings"})
		return
	}
	defer rows.Close()

	var bookings []FlightBooking
	for rows.Next() {
		var booking FlightBooking
		if err := rows.Scan(&booking.BookingID, &booking.FlightID, &booking.CustomerName, &booking.CustomerEmail,
			&booking.SeatNumber, &booking.BookingStatus, &booking.BookingTime); err != nil {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning booking data"})
			return
		}
		bookings = append(bookings, booking)
	}

	c.JSON(http.StatusOK, bookings)
}

func getBookingByID(c *gin.Context) {
	bookingID := c.Param("id")

	var booking FlightBooking
	err := db.QueryRow(context.Background(), "SELECT booking_id, flight_id, customer_name, customer_email, seats_booked, booking_status, booking_time FROM flight_bookings WHERE booking_id=$1", bookingID).Scan(
		&booking.BookingID, &booking.FlightID, &booking.CustomerName, &booking.CustomerEmail, &booking.SeatNumber, &booking.BookingStatus, &booking.BookingTime)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch booking"})
		}
		return
	}

	c.JSON(http.StatusOK, booking)
}

func deleteBooking(c *gin.Context) {
	bookingID := c.Param("id")

	_, err := db.Exec(context.Background(), "UPDATE flight_bookings SET booking_status=$1 WHERE booking_id=$2", BookingStatusCanceled, bookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking canceled successfully"})
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

	r.POST("/bookings", createBooking)
	r.GET("/bookings", getBookings)
	r.GET("/bookings/:id", getBookingByID)
	r.DELETE("/bookings/:id", deleteBooking)

	r.Run(":8080")
}
