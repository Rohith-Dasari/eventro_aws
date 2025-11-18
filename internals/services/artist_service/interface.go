package artistservice

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/artist_service_mock.go -package=mocks -source=interface.go
type ArtistServiceI interface {
	CreateArtist(ctx context.Context, name, bio string) error
	GetArtistByID(ctx context.Context, id string) (*models.ArtistDTO, error)
}
