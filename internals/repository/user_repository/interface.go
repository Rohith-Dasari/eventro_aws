package userrepository

import "eventro_aws/internals/models"

//go:generate mockgen -destination=../../mocks/user_repository_mock.go -package=mocks -source=interface.go
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	GetBlockedUsers() ([]models.User, error)
	GetUsers() ([]models.User, error)
}
