package artistservice

import (
	"context"
	"errors"
	"eventro_aws/internals/models"
	artistrepository "eventro_aws/internals/repository/artist_repository"

	"github.com/google/uuid"
)

type Artistservice struct {
	ArtistRepo artistrepository.ArtistRepositoryI
}

func NewArtistService(artistRepo artistrepository.ArtistRepositoryI) Artistservice {
	return Artistservice{
		ArtistRepo: artistRepo,
	}
}

func (as *Artistservice) CreateArtist(ctx context.Context, name, bio string) error {

	if len(bio) < 12 {
		return errors.New("bio must be at least 12 characters long")
	}

	artist := models.ArtistDTO{
		ArtistID: uuid.New().String(),
		Name:     name,
		Bio:      bio,
	}

	if err := as.ArtistRepo.Create(artist); err != nil {
		return err
	}

	return nil
}

func (as *Artistservice) GetArtistByID(ctx context.Context, id string) (*models.ArtistDTO, error) {
	return as.ArtistRepo.GetByID(id)
}
