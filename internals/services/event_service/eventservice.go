package eventservice

import (
	"context"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	"eventro_aws/internals/models"
	eventsrepository "eventro_aws/internals/repository/event_repository"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type EventService struct {
	EventRepo eventsrepository.EventRepositoryI
}

func NewEventService(eventRepo eventsrepository.EventRepositoryI) *EventService {
	return &EventService{EventRepo: eventRepo}
}

func (e *EventService) CreateNewEvent(ctx context.Context, name, description, duration string, category models.EventCategory, artistIDs []string) (models.EventResponse, error) {
	name = strings.ToLower(name)
	eventID := uuid.New().String()
	event := models.Event{
		ID:          eventID,
		Name:        name,
		Description: description,
		Duration:    duration,
		Category:    category,
		IsBlocked:   false,
		ArtistIDs:   artistIDs,
	}

	if err := e.EventRepo.Create(ctx, &event); err != nil {
		return models.EventResponse{}, fmt.Errorf("failed to create event: %w", err)
	}

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

func (s *EventService) BrowseEvents(ctx context.Context, city, name string, blocked bool) ([]*models.EventDTO, error) {
	var events []*models.EventDTO
	var err error

	if city != "" {
		events, err = s.EventRepo.GetEventsByCity(ctx, city)
		if err != nil {
			return nil, err
		}
	} else if name != "" {
		name = strings.ToLower(name)
		events, err = s.EventRepo.GetEventsByName(ctx, name)

		if err != nil {
			return nil, err
		}
	} else if blocked {
		events, err = s.EventRepo.GetBlockedEvents(ctx)
		if err != nil {
			return nil, err
		}
		return events, nil
	}

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil {
		return nil, err
	}

	if name != "" && role == "admin" {
		return events, nil
	}

	var blockedEvents []*models.EventDTO
	var unblockedEvents []*models.EventDTO
	if blocked {

		for _, event := range events {
			if event.IsBlocked {
				blockedEvents = append(blockedEvents, event)
			}
		}
		events = blockedEvents
	} else {
		for _, event := range events {
			if !event.IsBlocked {
				unblockedEvents = append(unblockedEvents, event)
			}
		}
		events = unblockedEvents
	}

	return events, nil
}

func (e *EventService) DeleteEvent(ctx context.Context, eventID string) error {
	if err := e.EventRepo.Delete(ctx, eventID); err != nil {
		return err
	}
	return nil
}

func (e *EventService) UpdateEvent(ctx context.Context, eventID string, isBlocked bool) error {
	// event, err := e.EventRepo.GetByID(ctx, eventID)
	// if err != nil {
	// 	return fmt.Errorf("error from GetByID" + err.Error())
	// }
	// event.IsBlocked = isBlocked

	if err := e.EventRepo.Update(ctx, eventID, isBlocked); err != nil {
		return err
	}

	return nil
}

func (e *EventService) GetHostEvents(ctx context.Context, hostID string) ([]*models.EventDTO, error) {
	return e.EventRepo.GetEventsHostedByHost(ctx, hostID)
}

func (s *EventService) GetEventByID(ctx context.Context, id string) (*models.EventDTO, error) {
	event, err := s.EventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("from get by id : " + err.Error())
	}

	return event, nil
}
