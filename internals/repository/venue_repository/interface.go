package venuerepository

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/venue_repository_mock.go -package=mocks -source=interface.go
type VenueRepositoryI interface {
	Create(ctx context.Context, venue *models.Venue) error
	GetByID(ctx context.Context, id string) (*models.VenueResponse, error)
	ListByHost(ctx context.Context, hostID string) ([]models.VenueResponse, error)
	Update(ctx context.Context, venueID string, isBlocked bool) error
	Delete(ctx context.Context, id string) error
}
