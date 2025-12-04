package showservice

import (
	"context"
	"eventro_aws/internals/models"
	"time"
)

type ShowServiceI interface {
	UpdateShow(ctx context.Context, showID string, userID string, isBlocked bool) error
	BrowseShows(ctx context.Context, eventID, city, date, venueID, hostID string) ([]models.ShowDTO, error)
	CreateShow(ctx context.Context, eventID string, venueID string,
		hostID string, price float64, showDate time.Time,
		showTime string) error
	GetShowByID(ctx context.Context, showID string) (*models.ShowDTO, error)
}
