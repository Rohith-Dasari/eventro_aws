package userservice

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/user_service_mock.go -package=mocks -source=interface.go
type UserServiceI interface {
	UpdateUser(ctx context.Context, userID string, req models.UpdateUserRequest) (models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	BrowseUsers(ctx context.Context, userID string, blocked *bool) ([]models.User, error)
	GetUserByMailID(ctx context.Context, mail string) (*models.User, error)
}
