package venuerepository

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/venue_repository_mock.go -package=mocks -source=interface.go
type VenueRepository interface {
	Create(ctx context.Context, venue *models.Venue) error
	GetByID(ctx context.Context, id string) (*models.Venue, error)
	ListByHost(ctx context.Context, hostID string) ([]models.Venue, error)
	Update(ctx context.Context, venue *models.Venue) error
	Delete(ctx context.Context, id string) error
}
