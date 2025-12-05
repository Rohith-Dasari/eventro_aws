package userservice

import (
	"context"
	"eventro_aws/internals/models"
	userrepository "eventro_aws/internals/repository/user_repository"
	"fmt"
)

type UserService struct {
	UserRepo userrepository.UserRepositoryI
}

func NewUserService(userRepo userrepository.UserRepositoryI) *UserService {
	return &UserService{UserRepo: userRepo}
}

func (s *UserService) GetUserByMailID(ctx context.Context, mail string) (*models.User, error) {
	if mail == "" {
		return nil, fmt.Errorf("invalid email")
	}

	user, err := s.UserRepo.GetByEmail(mail)
	if err != nil {
		return nil, err
	}
	return user, nil
}
