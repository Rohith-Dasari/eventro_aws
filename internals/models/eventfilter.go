package models

type EventFilter struct {
	EventID    string
	Name       string
	Category   string
	Location   string
	IsBlocked  *bool
	ArtistName string
}
