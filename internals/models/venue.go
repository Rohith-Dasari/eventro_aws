package models

type Venue struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" dynamodbav:"pk"`
	Name      string `gorm:"type:text;not null" dynamodbav:"venue_name"`
	HostID    string `gorm:"type:uuid;not null;index" dynamodbav:"sk"`
	City      string `gorm:"type:text;not null" dynamodbav:"venue_city"`
	State     string `gorm:"type:text;not null" dynamodbav:"venue_state"`
	IsBlocked bool   `gorm:"default:false" dynamodbav:"is_blocked"`
}

type VenueResponse struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" dynamodbav:"pk"`
	Name      string `gorm:"type:text;not null" dynamodbav:"venue_name"`
	HostID    string `gorm:"type:uuid;not null;index" dynamodbav:"sk"`
	City      string `gorm:"type:text;not null" dynamodbav:"venue_city"`
	State     string `gorm:"type:text;not null" dynamodbav:"venue_state"`
	IsBlocked bool   `gorm:"default:false" dynamodbav:"is_blocked"`
}

type VenueDTO struct {
	ID    string `json:"venue_id" dynamodbav:"pk"`
	Name  string `dynamodbav:"venue_name" json:"venue_name"`
	City  string `dynamodbav:"venue_city" json:"city"`
	State string `dynamodbav:"venue_state" json:"state"`
}
