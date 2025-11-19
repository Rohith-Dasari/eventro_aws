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
	UserEmail        string   `json:"user_email"`
	BookingDate      string   `json:"booking_date"`
	BookingID        string   `json:"booking_id"`
	ShowID           string   `json:"show_id"`
	TimeBooked       string   `json:"time_booked"`
	NumTicketsBooked int      `json:"num_tickets_booked"`
	TotalPrice       float64  `json:"total_price"`
	Seats            []string `json:"seats"`
	VenueCity        string   `json:"venue_city"`
	VenueName        string   `json:"venue_name"`
	VenueState       string   `json:"venue_state"`
	EventName        string   `json:"event_name"`
	EventDuration    string   `json:"event_duration"`
	EventID          string   `json:"event_id"`
}

type EventDTO struct {
	EventID     string   `dynamodbav:"pk" json:"id"`
	EventName   string   `dynamodbav:"event_name" json:"name"`
	Description string   `dynamodbav:"description" json:"description"`
	Duration    string   `dynamodbav:"duration" json:"duration"`
	Category    string   `dynamodbav:"category" json:"category"`
	IsBlocked   bool     `dynamodbav:"is_blocked" json:"is_blocked"`
	ArtistNames []string `dynamodbav:"artist_names" json:"artist_names"`
	ArtistIDs   []string `json:"artist_ids"`
}
