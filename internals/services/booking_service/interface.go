package bookingservice

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/booking_service_mock.go -package=mocks -source=interface.go
type BookingServiceInterface interface {
	AddBooking(ctx context.Context, userID string, showID string, seats []string) (*models.UserBookingDTO, error)
	BrowseBookings(ctx context.Context, bookingID string, userID string, showID string) ([]models.UserBookingDTO, error)
}
