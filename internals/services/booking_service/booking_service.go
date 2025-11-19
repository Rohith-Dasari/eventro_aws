package bookingservice

import (
	"context"
	"errors"
	"eventro_aws/internals/models"
	bookingrepository "eventro_aws/internals/repository/booking_repository"
	showrepository "eventro_aws/internals/repository/show_repository"
	"fmt"
	"regexp"
	"strings"
)

type BookingService struct {
	BookingRepo bookingrepository.BookingRepositoryI
	ShowRepo    showrepository.ShowRepositoryI
}

func NewBookingService(bRepo bookingrepository.BookingRepositoryI,
	sRepo showrepository.ShowRepositoryI) BookingService {
	return BookingService{
		BookingRepo: bRepo,
		ShowRepo:    sRepo,
	}
}

func (bs *BookingService) AddBooking(
	ctx context.Context,
	userID string,
	showID string,
	requestedSeats []string,
) (*models.UserBookingDTO, error) {
	show, err := bs.ShowRepo.GetByID(ctx, showID)
	if err != nil {
		return nil, fmt.Errorf("show not found: %w", err)
	}
	if show.IsBlocked {
		return nil, errors.New("cannot book tickets for a blocked show")
	}

	booked := make(map[string]bool)
	for _, s := range show.BookedSeats {
		booked[s] = true
	}

	for _, seat := range requestedSeats {
		if booked[seat] {
			return nil, fmt.Errorf("seat %s is already booked", seat)
		}
		if !bs.isValidTicket(seat, show.BookedSeats) {
			return nil, fmt.Errorf("seat %s is not valid", seat)
		}
	}

	//add time booked

	numTickets := len(requestedSeats)
	totalPrice := float64(numTickets) * show.Price
	newBooking := &models.Booking{
		UserID:            userID,
		ShowID:            showID,
		NumTickets:        numTickets,
		TotalBookingPrice: totalPrice,
		Seats:             requestedSeats,
	}

	if err := bs.BookingRepo.Create(ctx, newBooking); err != nil {
		return nil, fmt.Errorf("error creating booking: %w", err)
	}
	if err := bs.ShowRepo.UpdateShowBooking(ctx, models.Booking{ShowID: showID, Seats: requestedSeats}); err != nil {
		return nil, fmt.Errorf("failed to update show bookings: %w", err)
	}

	bookingDTO := models.UserBookingDTO{
		BookingID:        newBooking.BookingID,
		UserEmail:        newBooking.UserID,
		ShowID:           newBooking.ShowID,
		TimeBooked:       newBooking.TimeBooked.String(),
		NumTicketsBooked: newBooking.NumTickets,
		TotalPrice:       newBooking.TotalBookingPrice,
		Seats:            newBooking.Seats,
	}

	return &bookingDTO, nil
}

func (bs *BookingService) isValidTicket(userTicket string, bookedTickets []string) bool {
	userTicket = strings.ToUpper(userTicket)

	matched, err := regexp.MatchString(`^[A-J](10|[1-9])$`, userTicket)
	if err != nil || !matched {
		return false
	}

	for _, ticket := range bookedTickets {
		if strings.ToUpper(ticket) == userTicket {
			return false
		}
	}
	return true
}

func (bs *BookingService) BrowseBookings(ctx context.Context, userID string) ([]models.UserBookingDTO, error) {
	bookings, err := bs.BookingRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching bookings: %w", err)
	}

	return bookings, nil
}
