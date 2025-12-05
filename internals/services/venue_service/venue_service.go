package venueservice

import (
	"context"
	"eventro_aws/internals/models"
	venuerepository "eventro_aws/internals/repository/venue_repository"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type VenueService struct {
	VenueRepo venuerepository.VenueRepositoryI
}

func NewVenueService(repo venuerepository.VenueRepositoryI) *VenueService {
	return &VenueService{VenueRepo: repo}
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
		ID:     venueID,
		HostID: hostID,
		Name:   name,
		City:   city,
		State:  state,
	}

	return venueDTO, nil
}

func (s *VenueService) UpdateVenue(ctx context.Context, venueID, userID, userRole string, isBlocked bool) error {
	venue, err := s.VenueRepo.GetByID(ctx, venueID)
	if err != nil {
		return fmt.Errorf("error from get venue by id %s: ", err.Error())
	}
	if strings.ToLower(userRole) != "admin" {
		if venue.HostID != userID {
			return fmt.Errorf("forbidden: cannot update another user's venue")
		}
	}
	if err := s.VenueRepo.Update(ctx, venueID, isBlocked); err != nil {
		return err
	}

	return nil
}

func (s *VenueService) DeleteVenue(ctx context.Context, venueID, userID, userRole string) error {
	venue, err := s.VenueRepo.GetByID(ctx, venueID)
	if err != nil {
		return err
	}

	if venue.HostID != userID && strings.ToLower(userRole) != "admin" {
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
	return venues, nil

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
