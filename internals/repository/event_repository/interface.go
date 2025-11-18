package eventrepository

import (
	"context"
	"eventro_aws/internals/models"
)

//go:generate mockgen -destination=../../mocks/event_repository_mock.go -package=mocks -source=interface.go
type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, eventID string) (*models.EventDTO, error)
	Update(ctx context.Context, eventID string, isBlocked bool) error
	Delete(ctx context.Context, id string) error
	GetEventsByCity(ctx context.Context, city string) ([]models.EventDTO, error)
	GetEventsHostedByHost(ctx context.Context, hostID string) ([]models.EventDTO, error)
}
