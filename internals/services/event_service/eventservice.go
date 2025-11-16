package eventservice

import (
	"context"
	"eventro_aws/internals/models"
	eventsrepository "eventro_aws/internals/repository/event_repository"
	"fmt"

	"github.com/google/uuid"
)

type EventService struct {
	EventRepo eventsrepository.EventRepository
}

func NewEventService(eventRepo eventsrepository.EventRepository) EventService {
	return EventService{EventRepo: eventRepo}
}

func (e *EventService) CreateNewEvent(ctx context.Context, name, description, duration string, category models.EventCategory, artistIDs []string) (models.EventResponse, error) {
	eventID := uuid.New().String()
	// Create the event
	event := models.Event{
		ID:          eventID,
		Name:        name,
		Description: description,
		Duration:    duration,
		Category:    category,
		IsBlocked:   false,
	}

	if err := e.EventRepo.Create(&event); err != nil {
		return models.EventResponse{}, fmt.Errorf("failed to create event: %w", err)
	}

	// for _, artistID := range artistIDs {
	// 	eventArtist := models.EventArtist{
	// 		EventID:  eventID,
	// 		ArtistID: artistID,
	// 	}
	// 	if err := e.EventRepo.AddEventArtist(&eventArtist); err != nil {
	// 		return models.EventResponse{}, fmt.Errorf("failed to associate artist %s: %w", artistID, err)
	// 	}
	// }

	return models.EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Duration:    event.Duration,
		Category:    string(event.Category),
		IsBlocked:   event.IsBlocked,
		ArtistIDs:   artistIDs,
	}, nil
}

func (s *EventService) BrowseEvents(ctx context.Context, filter models.EventFilter) ([]models.EventResponse, error) {
	events, err := s.EventRepo.GetFilteredEvents(filter)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (e *EventService) DeleteEvent(ctx context.Context, eventID string) error {
	if err := e.EventRepo.Delete(eventID); err != nil {
		return err
	}
	return nil
}

func (e *EventService) UpdateEvent(ctx context.Context, eventID string, updateData models.EventUpdate) (models.EventResponse, error) {
	// Fetch existing event
	event, err := e.EventRepo.GetByID(eventID)
	if err != nil {
		return models.EventResponse{}, err
	}

	if updateData.Name != nil {
		event.Name = *updateData.Name
	}
	if updateData.Description != nil {
		event.Description = *updateData.Description
	}
	if updateData.Duration != nil {
		event.Duration = *updateData.Duration
	}
	if updateData.Category != nil {
		event.Category = *updateData.Category
	}
	if updateData.IsBlocked != nil {
		event.IsBlocked = *updateData.IsBlocked
	}

	// Save updated event
	if err := e.EventRepo.Update(event); err != nil {
		return models.EventResponse{}, err
	}

	return models.EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		Duration:    event.Duration,
		Category:    string(event.Category),
		IsBlocked:   event.IsBlocked,
	}, nil
}

func (e *EventService) GetHostEvents(ctx context.Context, hostID string) ([]models.EventResponse, error) {
	return e.EventRepo.GetEventsHostedByHost(hostID)
}
