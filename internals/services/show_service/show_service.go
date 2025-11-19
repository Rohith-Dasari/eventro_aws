package showservice

import (
	"context"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	"eventro_aws/internals/models"
	showrepository "eventro_aws/internals/repository/show_repository"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ShowService struct {
	ShowRepo showrepository.ShowRepositoryI
}

func NewShowService(
	showRepo showrepository.ShowRepositoryI,
) ShowService {
	return ShowService{
		ShowRepo: showRepo,
	}
}

func (s *ShowService) UpdateShow(ctx context.Context, showID string, userID string, isBlocked bool) error {
	// fetch show first
	_, err := s.ShowRepo.GetByID(ctx, showID)
	if err != nil {
		return err
	}

	// authorization check (only host and admin can update)
	userRole, _ := authenticationmiddleware.GetUserRole(ctx)
	if strings.ToLower(userRole) != "admin" && strings.ToLower(userRole) != "host" {
		return fmt.Errorf("forbidden: cannot update another user's show")
	}

	if err := s.ShowRepo.Update(ctx, showID, isBlocked); err != nil {
		return err
	}

	return nil
}

func (s *ShowService) BrowseShows(ctx context.Context, eventID, city, date, venueID, hostID string) ([]models.ShowDTO, error) {
	shows, err := s.ShowRepo.ListByEvent(ctx, eventID, city, date, venueID, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shows: %w", err)
	}

	if hostID != "" {
		//only keep shows which has shows[i].HostID =hostID
		filtered := make([]models.ShowDTO, 0, len(shows))
		for _, show := range shows {
			if show.HostID == hostID {
				filtered = append(filtered, show)
			}
		}
		return filtered, nil
	}
	return shows, nil
}

func (s *ShowService) CreateShow(ctx context.Context, eventID string, venueID string,
	hostID string, price float64, showDate time.Time,
	showTime string) error {
	showID := uuid.New().String()

	show := models.Show{
		ID:          showID,
		HostID:      hostID,
		VenueID:     venueID,
		EventID:     eventID,
		IsBlocked:   false,
		Price:       price,
		ShowDate:    showDate,
		ShowTime:    showTime,
		BookedSeats: []string{},
	}

	if err := s.ShowRepo.Create(ctx, &show); err != nil {
		return fmt.Errorf("failed to create show: %w", err)
	}

	return nil
}
