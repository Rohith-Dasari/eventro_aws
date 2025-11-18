package models

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type SignupRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type SignupResponse struct {
	Token string `json:"token"`
}

type ArtistDTO struct {
	Name     string `dynamodbav:"sk"`
	ArtistID string `dynamodbav:"pk"`
	Bio      string `dynamodbav:"bio"`
}

type UserBookingDTO struct {
	UserEmail        string
	BookingDate      string
	BookingID        string
	ShowID           string
	TimeBooked       string
	NumTicketsBooked int
	TotalPrice       float64
	Seats            []string
	VenueCity        string
	VenueName        string
	VenueState       string
	EventName        string
	EventDuration    string
	EventID          string
}

type EventDTO struct {
	EventID     string   `dynamodbav:"pk"`
	EventName   string   `dynamodbav:"event_name"`
	Description string   `dynamodbav:"description"`
	Duration    string   `dynamodbav:"duration"`
	Category    string   `dynamodbav:"category"`
	IsBlocked   bool     `dynamodbav:"is_blocked"`
	ArtistNames []string `dynamodbav:"artist_names"`
	ArtistBios  []string `dynamodbav:"artist_bios"`
}
