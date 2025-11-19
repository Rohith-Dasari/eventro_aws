package venueservice

import (
	"context"
	"eventro_aws/internals/models"
	venuerepository "eventro_aws/internals/repository/venue_repository"
	"fmt"

	"github.com/google/uuid"
)

type VenueService struct {
	VenueRepo venuerepository.VenueRepository
}

func NewVenueService(repo venuerepository.VenueRepository) VenueService {
	return VenueService{VenueRepo: repo}
}

func (vs *VenueService) CreateVenue(ctx context.Context, hostID, name, city, state string, isSeatLayoutRequired bool) (models.VenueResponse, error) {
	venueID := uuid.New().String()

	venue := models.Venue{
		ID:     venueID,
		HostID: hostID,
		Name:   name,
		City:   city,
		State:  state,
	}

	if err := vs.VenueRepo.Create(ctx, &venue); err != nil {
		return models.VenueResponse{}, fmt.Errorf("failed to create venue: %w", err)
	}
	venueDTO := models.VenueResponse{
		ID:                   venueID,
		HostID:               hostID,
		Name:                 name,
		City:                 city,
		State:                state,
		IsSeatLayoutRequired: isSeatLayoutRequired,
	}

	return venueDTO, nil
}

func (s *VenueService) UpdateVenue(ctx context.Context, venueID, userID, userRole string, update models.UpdateVenueData) (models.VenueResponse, error) {
	venue, err := s.VenueRepo.GetByID(ctx, venueID)
	if err != nil {
		return models.VenueResponse{}, err
	}

	if venue.HostID != userID && update.IsBlocked == nil {
		return models.VenueResponse{}, fmt.Errorf("forbidden: cannot update another user's venue")
	}

	if update.Name != nil {
		venue.Name = *update.Name
	}
	if update.City != nil {
		venue.City = *update.City
	}
	if update.State != nil {
		venue.State = *update.State
	}

	if update.IsBlocked != nil {
		if venue.HostID == userID || userRole == "admin" {
			venue.IsBlocked = *update.IsBlocked
		} else {
			return models.VenueResponse{}, fmt.Errorf("forbidden: only host or admin can block/unblock")
		}
	}

	if err := s.VenueRepo.Update(ctx, venue); err != nil {
		return models.VenueResponse{}, err
	}
	venueDTO := models.VenueResponse{
		ID:        venue.ID,
		HostID:    venue.HostID,
		Name:      venue.Name,
		City:      venue.City,
		State:     venue.State,
		IsBlocked: venue.IsBlocked,
	}

	return venueDTO, nil
}

func (s *VenueService) DeleteVenue(ctx context.Context, venueID, userID, userRole string) error {
	venue, err := s.VenueRepo.GetByID(ctx, venueID)
	if err != nil {
		return err
	}

	if venue.HostID != userID && userRole != "admin" {
		return fmt.Errorf("forbidden: cannot delete this venue")
	}

	if err := s.VenueRepo.Delete(ctx, venueID); err != nil {
		return err
	}

	return nil
}

func (s *VenueService) GetHostVenues(ctx context.Context, hostID string) ([]models.VenueResponse, error) {
	venues, err := s.VenueRepo.ListByHost(ctx, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venues: %w", err)
	}
	venueDTO := make([]models.VenueResponse, len(venues))
	for i, v := range venues {
		venueDTO[i] = models.VenueResponse{
			ID:        v.ID,
			HostID:    v.HostID,
			Name:      v.Name,
			City:      v.City,
			State:     v.State,
			IsBlocked: v.IsBlocked,
		}
	}
	return venueDTO, nil

}
func (s *VenueService) GetVenueByID(ctx context.Context, venueID string) (*models.VenueResponse, error) {
	v, err := s.VenueRepo.GetByID(ctx, venueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venues: %w", err)
	}
	venueDTO := models.VenueResponse{
		ID:        v.ID,
		HostID:    v.HostID,
		Name:      v.Name,
		City:      v.City,
		State:     v.State,
		IsBlocked: v.IsBlocked,
	}
	return &venueDTO, nil

}
