package venueservice

import (
	"context"
	"eventro_aws/internals/models"
)

type VenueServiceI interface {
	CreateVenue(ctx context.Context, hostID, name, city, state string, isSeatLayoutRequired bool) (models.VenueResponse, error)
	UpdateVenue(ctx context.Context, venueID, userID, userRole string, isBlocked bool) error
	DeleteVenue(ctx context.Context, venueID, userID, userRole string) error
	GetHostVenues(ctx context.Context, hostID string) ([]models.VenueResponse, error)
	GetVenueByID(ctx context.Context, venueID string) (*models.VenueResponse, error)
}
