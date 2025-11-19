package showrepository

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/show_repository_mock.go -package=mocks -source=interface.go
type ShowRepositoryI interface {
	Create(ctx context.Context, show *models.Show) error
	GetByID(ctx context.Context, id string) (*models.ShowDTO, error)
	ListByEvent(ctx context.Context, eventID, city, date, venueID, hostID string) ([]models.ShowDTO, error)
	Update(ctx context.Context, showID string, isBlocked bool) error
	UpdateShowBooking(ctx context.Context, booking models.Booking) error
}
