package userrepository

import (
	"eventro_aws/internals/models"

	"gorm.io/gorm"
)



type UserRepositoryPG struct {
	db *gorm.DB
}

// constructor
func NewUserRepositoryPG(db *gorm.DB) *UserRepositoryPG {
	return &UserRepositoryPG{db: db}
}

// Create a new user
func (r *UserRepositoryPG) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Get user by ID
func (r *UserRepositoryPG) GetByID(id string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "user_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Get user by Email
func (r *UserRepositoryPG) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List all users
func (r *UserRepositoryPG) GetUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Update user
func (r *UserRepositoryPG) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete user
func (r *UserRepositoryPG) Delete(id string) error {
	return r.db.Delete(&models.User{}, "user_id = ?", id).Error
}

func (r *UserRepositoryPG) GetBlockedUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("is_blocked = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
