package eventservice

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/event_service_mock.go -package=mocks -source=interface.go
type EventServiceI interface {
	CreateNewEvent(ctx context.Context, name, description, duration string, category models.EventCategory, artistIDs []string) (models.EventResponse, error)
	BrowseEvents(ctx context.Context, filter models.EventFilter) ([]models.EventResponse, error)
	DeleteEvent(ctx context.Context, eventID string) error
	UpdateEvent(ctx context.Context, eventID string, updateData models.EventUpdate) (models.EventResponse, error)
	GetHostEvents(ctx context.Context, hostID string) ([]models.EventResponse, error)
}
