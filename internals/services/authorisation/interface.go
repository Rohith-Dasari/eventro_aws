package authorisation

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/auth_service_mock.go -package=mocks -source=interface.go

type AuthServiceI interface {
	ValidateLogin(ctx context.Context, email, password string) (models.User, error)
	Signup(ctx context.Context, username, email, phoneNumber, password string) (models.User, error)
}
