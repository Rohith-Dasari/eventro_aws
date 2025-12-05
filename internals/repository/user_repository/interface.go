package userrepository

import "eventro_aws/internals/models"

//go:generate mockgen -destination=../../mocks/user_repository_mock.go -package=mocks -source=interface.go
type UserRepositoryI interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
}
