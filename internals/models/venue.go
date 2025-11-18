package models

type Venue struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" dynamodbav:"pk"`
	Name      string `gorm:"type:text;not null" dynamodbav:"name"`
	HostID    string `gorm:"type:uuid;not null;index" dynamodbav:"sk"`
	City      string `gorm:"type:text;not null" dynamodbav:"city"`
	State     string `gorm:"type:text;not null" dynamodbav:"state"`
	IsBlocked bool   `gorm:"default:false" dynamodbav:"is_blocked"`
}

type VenueResponse struct {
	ID                   string
	Name                 string
	HostID               string
	City                 string
	State                string
	IsSeatLayoutRequired bool
	IsBlocked            bool
}

type VenueDTO struct {
	ID    string `json:"id"`
	Name  string `dynamodbav:"venue_name"`
	City  string `dynamodbav:"venue_city"`
	State string `dynamodbav:"venue_state"`
}
