package models

type Role string

const (
	Admin    Role = "Admin"
	Host     Role = "Host"
	Customer Role = "Customer"
)

type User struct {
	UserID      string `dynamodbav:"user_id"`
	Username    string `dynamodbav:"username"`
	Email       string `dynamodbav:"pk"`
	PhoneNumber string `dynamodbav:"phone_number"`
	Password    string `dynamodbav:"password"`
	Role        Role   `dynamodbav:"role"`
	IsBlocked   bool   `dynamodbav:"is_blocked"`
}

type UpdateUserRequest struct {
	IsBlocked *bool   `json:"isBlocked,omitempty"`
	Role      *string `json:"role,omitempty"`
}
