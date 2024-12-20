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

type Hotel struct {
	HotelID        string `json:"hotelId" db:"hotel_id"`
	Name           string `json:"name" db:"name"`
	Location       string `json:"location" db:"location"`
	Rating         int    `json:"rating" db:"rating"`
	RoomsAvailable int    `json:"roomsAvailable" db:"rooms_available"`
}

type HotelBooking struct {
	BookingID     string        `json:"bookingId" db:"booking_id"`
	HotelID       string        `json:"hotelId" db:"hotel_id"`
	CheckInDate   time.Time     `json:"checkInDate" db:"check_in_date"`
	CheckOutDate  time.Time     `json:"checkOutDate" db:"check_out_date"`
	CustomerName  string        `json:"customerName" db:"customer_name"`
	CustomerEmail string        `json:"customerEmail" db:"customer_email"`
	RoomNumber    int           `json:"roomNumber" db:"room_number"`
	BookingStatus BookingStatus `json:"bookingStatus" db:"booking_status"`
	BookingTime   time.Time     `json:"bookingTime" db:"booking_time"`
	UpdatedAt     time.Time     `json:"updatedAt" db:"updated_at"`
}

// ####################################################################################################################################
// ### Database 																																																										###
// ####################################################################################################################################

var db *pgxpool.Pool

func connectDB() {
	databaseURL := "host=localhost user=saga-hotel password=saga-hotel dbname=saga-hotel sslmode=disable port=5433" // e.g., "postgres://user:password@localhost:5432/dbname"
	// databaseURL := "host=database-hotel user=saga-hotel password=saga-hotel dbname=saga-hotel sslmode=disable port=5433" // e.g., "postgres://user:password@localhost:5432/dbname"
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
// ### 	Request Objects 																																																						###
// ####################################################################################################################################

type CreateHotelRequest struct {
	Name           string `json:"name" binding:"required"`
	Location       string `json:"location" binding:"required"`
	Rating         int    `json:"rating"`
	RoomsAvailable int    `json:"roomsAvailable" binding:"required"`
}

type CreateHotelBookingRequest struct {
	HotelID       string    `json:"hotelId" binding:"required"`
	CheckInDate   time.Time `json:"checkInDate" binding:"required"`
	CheckOutDate  time.Time `json:"checkOutDate" binding:"required"`
	CustomerName  string    `json:"customerName" binding:"required"`
	CustomerEmail string    `json:"customerEmail" binding:"required"`
	RoomNumber    int       `json:"roomNumber" binding:"required"`
}

// ####################################################################################################################################
// ### Handlers 																																																										###
// ####################################################################################################################################

// Hotel
func createHotel(c *gin.Context) {
	var req CreateHotelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newHotel := Hotel{
		HotelID:        uuid.New().String(),
		Name:           req.Name,
		Location:       req.Location,
		Rating:         req.Rating,
		RoomsAvailable: req.RoomsAvailable,
	}

	_, err := db.Exec(context.Background(),
		"INSERT INTO hotels (hotel_id, name, location, rating, rooms_available) VALUES ($1, $2, $3, $4, $5)",
		newHotel.HotelID, newHotel.Name, newHotel.Location, newHotel.Rating, newHotel.RoomsAvailable)

	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hotel"})
		return
	}

	c.JSON(http.StatusCreated, newHotel)
}

func getHotels(c *gin.Context) {
	rows, err := db.Query(context.Background(),
		"SELECT hotel_id, name, location, rating, rooms_available FROM hotels")
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch hotels"})
		return
	}
	defer rows.Close()

	var hotels []Hotel
	for rows.Next() {
		var hotel Hotel
		if err := rows.Scan(&hotel.HotelID, &hotel.Name, &hotel.Location, &hotel.Rating, &hotel.RoomsAvailable); err != nil {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning hotel data"})
			return
		}
		hotels = append(hotels, hotel)
	}

	c.JSON(http.StatusOK, hotels)
}

func getHotelByID(c *gin.Context) {
	hotelID := c.Param("id")

	var hotel Hotel
	err := db.QueryRow(context.Background(),
		"SELECT hotel_id, name, location, rating, rooms_available FROM hotels WHERE hotel_id=$1", hotelID).
		Scan(&hotel.HotelID, &hotel.Name, &hotel.Location, &hotel.Rating, &hotel.RoomsAvailable)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Hotel not found"})
		} else {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hotel"})
		}
		return
	}

	c.JSON(http.StatusOK, hotel)
}

func deleteFlight(c *gin.Context) {
	hotelID := c.Param("id")

	_, err := db.Exec(context.Background(),
		"DELETE FROM hotels WHERE hotel_id=$1",
		hotelID)
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete hotel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Hotel deleted successfully"})
}

// HotelBooking
func createHotelBooking(c *gin.Context) {
	var req CreateHotelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	println(req.HotelID)
	println(req.CustomerName)
	println(req.CustomerEmail)
	println(req.RoomNumber)

	var hotelExists bool
	err := db.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM hotels WHERE hotel_id=$1)", req.HotelID).
		Scan(&hotelExists)
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate hotel"})
		return
	}

	if !hotelExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hotel does not exist"})
		return
	}

	newBooking := HotelBooking{
		BookingID:     uuid.New().String(),
		HotelID:       req.HotelID,
		CheckInDate:   req.CheckInDate,
		CheckOutDate:  req.CheckOutDate,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		RoomNumber:    req.RoomNumber,
		BookingStatus: BookingStatusActive,
		BookingTime:   time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = db.Exec(context.Background(),
		"INSERT INTO hotel_bookings (booking_id, hotel_id, check_in_date, check_out_date, customer_name, customer_email, room_number, booking_status, booking_time, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		newBooking.BookingID, newBooking.HotelID, newBooking.CheckInDate, newBooking.CheckOutDate, newBooking.CustomerName, newBooking.CustomerEmail, newBooking.RoomNumber, newBooking.BookingStatus, newBooking.BookingTime, newBooking.UpdatedAt)

	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, newBooking)
}

func getHotelBookings(c *gin.Context) {
	bookingStatus := c.Query("bookingStatus")
	hotelID := c.Query("hotelId")
	bookingID := c.Query("bookingId")
	customerEmail := c.Query("customerEmail")

	query := "SELECT booking_id, hotel_id, check_in_date, check_out_date, customer_name, customer_email, room_number, booking_status FROM hotel_bookings WHERE 1=1"
	args := []interface{}{}
	argID := 1

	if bookingStatus != "" {
		query += " AND booking_status=$" + strconv.Itoa(argID)
		args = append(args, bookingStatus)
		argID++
	}
	if hotelID != "" {
		query += " AND hotel_id=$" + strconv.Itoa(argID)
		args = append(args, hotelID)
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

	var bookings []HotelBooking
	for rows.Next() {
		var booking HotelBooking
		if err := rows.Scan(&booking.BookingID, &booking.HotelID, &booking.CheckInDate, &booking.CheckOutDate, &booking.CustomerName, &booking.CustomerEmail, &booking.RoomNumber, &booking.BookingStatus); err != nil {
			println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning booking data"})
			return
		}
		bookings = append(bookings, booking)
	}

	c.JSON(http.StatusOK, bookings)
}

func getHotelBookingByID(c *gin.Context) {
	bookingID := c.Param("id")

	var booking HotelBooking
	err := db.QueryRow(context.Background(),
		"SELECT booking_id, hotel_id, check_in_date, check_out_date, customer_name, customer_email, room_number, booking_status FROM hotel_bookings WHERE booking_id=$1", bookingID).
		Scan(&booking.BookingID, &booking.HotelID, &booking.CheckInDate, &booking.CheckOutDate, &booking.CustomerName, &booking.CustomerEmail, &booking.RoomNumber, &booking.BookingStatus)

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
		c.JSON(http.StatusBadRequest,  gin.H{"error": "Failed to cancel (Missing bookingId)"})
		return
	}

	println(bookingID)

	_, err := db.Exec(context.Background(),
		"UPDATE hotel_bookings SET booking_status=$1, updated_at=$2 WHERE booking_id=$3",
		BookingStatusCanceled, time.Now(), bookingID)
	if err != nil {
		println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking canceled successfully"})
}

func deleteBookingByHotelRoom(c *gin.Context) {
	hotelID := c.Param("hotelId")
	roomNumber := c.Param("roomNumber")

	if hotelID == "" {
		c.JSON(http.StatusBadRequest,  gin.H{"error": "Failed to cancel (Missing hotelId)"})
		return
	}

	if roomNumber == "" {
		c.JSON(http.StatusBadRequest,  gin.H{"error": "Failed to cancel (Missing roomNumber)"})
		return
	}

	println(hotelID)
	println(roomNumber)

	_, err := db.Exec(context.Background(),
		"UPDATE hotel_bookings SET booking_status=$1, updated_at=$2 WHERE hotel_id=$3 AND room_number=$4",
		BookingStatusCanceled, time.Now(), hotelID, roomNumber)
	if err != nil {
		println(err.Error())
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

	r.POST("/hotels", createHotel)
	r.GET("/hotels", getHotels)
	r.GET("/hotels/:id", getHotelByID)
	r.DELETE("/hotels/:id", deleteFlight)

	r.POST("/bookings", createHotelBooking)
	r.GET("/bookings", getHotelBookings)
	r.GET("/bookings/:id", getHotelBookingByID)
	r.DELETE("/bookings/id/:id", deleteBookingByID)
	r.DELETE("/bookings/room/:hotelId/:roomBNumber", deleteBookingByHotelRoom)

	r.Run(":3001")
}
