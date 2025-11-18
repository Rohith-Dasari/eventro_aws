package artistrepository

import "eventro_aws/internals/models"

type ArtistRepositoryI interface {
	Create(artist models.ArtistDTO) error
	GetByID(id string) (*models.ArtistDTO, error)
}
