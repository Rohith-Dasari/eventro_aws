package eventservice

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/event_service_mock.go -package=mocks -source=interface.go
type EventServiceI interface {
	CreateNewEvent(ctx context.Context, name, description, duration string, category models.EventCategory, artistIDs []string) (models.EventResponse, error)
	BrowseEvents(ctx context.Context, city, name string, blocked bool) ([]*models.EventDTO, error)
	DeleteEvent(ctx context.Context, eventID string) error
	UpdateEvent(ctx context.Context, eventID string, isBlocked bool) error
	GetHostEvents(ctx context.Context, hostID string) ([]models.EventDTO, error)
	GetEventByID(ctx context.Context, id string) (*models.EventDTO, error)
}
