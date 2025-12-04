package bookingservice

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/booking_service_mock.go -package=mocks -source=interface.go
type BookingServiceI interface {
	AddBooking(
		ctx context.Context,
		userID string,
		showID string,
		requestedSeats []string,
	) (*models.UserBookingDTO, error)
	BrowseBookings(ctx context.Context, userID string) ([]models.UserBookingDTO, error)
}
