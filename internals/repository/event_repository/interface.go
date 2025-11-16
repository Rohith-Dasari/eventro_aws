package eventrepository

import "eventro_aws/internals/models"

//go:generate mockgen -destination=../../mocks/event_repository_mock.go -package=mocks -source=interface.go
type EventRepository interface {
	Create(event *models.Event) error
	GetByID(id string) (*models.Event, error)
	List() ([]models.Event, error)
	Update(event *models.Event) error
	Delete(id string) error
	AddEventArtist(ea *models.EventArtist) error
	GetArtistsByEventID(eventID string) ([]models.Artist, error)
	GetEventsByCity(city string) ([]models.Event, error)
	GetFilteredEvents(filter models.EventFilter) ([]models.EventResponse, error)
	GetEventsHostedByHost(hostID string) ([]models.EventResponse, error)
}
