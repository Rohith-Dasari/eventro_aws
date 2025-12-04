package userservice

import (
	"context"
	"errors"
	"eventro_aws/internals/models"
	userrepository "eventro_aws/internals/repository/user_repository"
	"fmt"

	"gorm.io/gorm"
)

type UserService struct {
	UserRepo userrepository.UserRepositoryI
}

func NewUserService(userRepo userrepository.UserRepositoryI) *UserService {
	return &UserService{UserRepo: userRepo}
}

func (s *UserService) BrowseUsers(ctx context.Context, userID string, blocked *bool) ([]models.User, error) {
	if blocked != nil && *blocked {
		return s.UserRepo.GetBlockedUsers()
	}

	return s.UserRepo.GetUsers()
}

func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("invalid user ID")
	}

	user, err := s.UserRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return user, nil
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

func (s *UserService) UpdateUser(ctx context.Context, userID string, req models.UpdateUserRequest) (models.User, error) {
	user, err := s.UserRepo.GetByID(userID)
	if err != nil {
		return models.User{}, err
	}

	if req.IsBlocked != nil {
		user.IsBlocked = *req.IsBlocked
	}

	if req.Role != nil {
		user.Role = models.Role(*req.Role)
	}

	if err := s.UserRepo.Update(user); err != nil {
		return models.User{}, err
	}

	return *user, nil
}
