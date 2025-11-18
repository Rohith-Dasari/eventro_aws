package bookingrepository

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/booking_repository_mock.go -package=mocks -source=interface.go
type BookingRepositoryI interface {
	Create(ctx context.Context, booking *models.Booking) error
	ListByUser(ctx context.Context, userID string) ([]models.UserBookingDTO, error)
}
