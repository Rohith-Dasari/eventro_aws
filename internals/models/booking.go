package models

import (
	"time"

	"github.com/lib/pq"
)

type Booking struct {
	BookingID         string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID            string         `gorm:"type:uuid;not null"`
	User              User           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ShowID            string         `gorm:"type:uuid;not null;index"`
	Show              Show           `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TimeBooked        time.Time      `gorm:"autoCreateTime"`
	NumTickets        int            `gorm:"not null"`
	TotalBookingPrice float64        `gorm:"not null"`
	Seats             pq.StringArray `gorm:"type:text[]"`
}


// trigger showid search, you get venue id and event id, along with others, trigger venue id and event id search to get other fields
//update show booked seats
// successful creation of booking