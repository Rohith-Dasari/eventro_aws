package authorisation

import (
	"context"
	"errors"
	"eventro_aws/internals/models"
	userrepository "eventro_aws/internals/repository/user_repository"
	"net/mail"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo userrepository.UserRepositoryI
}

func NewAuthService(userRepo userrepository.UserRepositoryI) *AuthService {
	return &AuthService{
		UserRepo: userRepo,
	}
}

func (a *AuthService) ValidateLogin(ctx context.Context, email, password string) (models.User, error) {
	user, err := a.UserRepo.GetByEmail(email)
	if err != nil {
		return models.User{}, err
	}
	if user.IsBlocked {
		return models.User{}, errors.New("user account is blocked, please contact admin")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return models.User{}, errors.New("invalid email or password")
	}

	return *user, nil
}

func (a *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

func (a *AuthService) IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (a *AuthService) IsValidPhoneNumber(phone string) bool {
	var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{9,14}$`)
	return phoneRegex.MatchString(phone)
}

func (a *AuthService) IsValidPassword(password string) bool {
	var (
		minLength = len(password) >= 12
		hasNumber = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasUpper  = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower  = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasSymbol = regexp.MustCompile(`[!@#$%^&*()\-+]`).MatchString(password)
	)

	return minLength && hasNumber && hasUpper && hasLower && hasSymbol
}

func (a *AuthService) Signup(ctx context.Context, username, email, phoneNumber, password string) (models.User, error) {
	if !a.IsValidEmail(email) {
		return models.User{}, errors.New("invalid email format")
	}

	if !a.IsValidPassword(password) {
		return models.User{}, errors.New("password must be at least 12 characters long, and include uppercase, lowercase, number, and symbol")
	}

	hashedPassword, err := a.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	if !a.IsValidPhoneNumber(phoneNumber) {
		return models.User{}, errors.New("invalid phone number format")
	}

	newUser := models.User{
		UserID:      uuid.New().String(),
		Username:    username,
		Email:       email,
		PhoneNumber: phoneNumber,
		Password:    hashedPassword,
		Role:        models.Customer,
	}

	if err := a.UserRepo.Create(&newUser); err != nil {
		return models.User{}, err
	}

	return newUser, nil
}
