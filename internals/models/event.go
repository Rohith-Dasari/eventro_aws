package models

import "github.com/lib/pq"

type EventCategory string

const (
	Movie    EventCategory = "movie"
	Sports   EventCategory = "sports"
	Concert  EventCategory = "concert"
	Workshop EventCategory = "workshop"
	Party    EventCategory = "party"
)

type Event struct {
	ID          string        `json:"id" dynamodbav:"pk"`
	Name        string        `json:"name" dynamodbav:"event_name"`
	Description string        `json:"description" dynamodbav:"description"`
	Duration    string        `json:"duration" dynamodbav:"duration"`
	Category    EventCategory `json:"category" dynamodbav:"category"`
	IsBlocked   bool          `json:"is_blocked" dynamodbav:"is_blocked"`
	ArtistIDs   []string      `json:"artist_ids,omitempty" dynamodbav:"artist_ids"`
}

type EventArtist struct {
	EventID  string `gorm:"primaryKey;type:uuid"`
	ArtistID string `gorm:"primaryKey;type:uuid"`

	Event  Event  `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE"`
	Artist Artist `gorm:"foreignKey:ArtistID;references:ID;constraint:OnDelete:CASCADE"`
}

type EventResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Duration    string         `json:"duration"`
	Category    string         `json:"category"`
	IsBlocked   bool           `json:"is_blocked"`
	ArtistIDs   pq.StringArray `json:"artist_ids" gorm:"type:text[]"`
	ArtistNames pq.StringArray `json:"artist_names" gorm:"type:text[]"`
}
