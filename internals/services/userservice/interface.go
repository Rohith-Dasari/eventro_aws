package userservice

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/user_service_mock.go -package=mocks -source=interface.go
type UserServiceI interface {
	GetUserByMailID(ctx context.Context, mail string) (*models.User, error)
}
