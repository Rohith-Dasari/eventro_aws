package main

import (
	"context"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	bookingrepository "eventro_aws/internals/repository/booking_repository"
	showrepository "eventro_aws/internals/repository/show_repository"
	bookingservice "eventro_aws/internals/services/booking_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var bookingService bookingservice.BookingService

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	bookingRepo := bookingrepository.NewBookingRepositoryDDB(ddb, "eventro")
	showRepo := showrepository.NewShowRepositoryDDB(ddb, "eventro")
	bookingService = bookingservice.NewBookingService(bookingRepo, showRepo)
}

func main() {
	lambda.Start(GetBookingsOfUser)
}

func GetBookingsOfUser(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	authUserID, _ := authenticationmiddleware.GetUserEmail(ctx)
	role, _ := authenticationmiddleware.GetUserRole(ctx)

	userID := event.QueryStringParameters["userId"]

	if role != "admin" {
		userID = authUserID
	}

	bookings, err := bookingService.BrowseBookings(ctx, userID)
	if err != nil {
		return customresponse.SendCustomResponse(500, "failed to fetch bookings")
	}

	return customresponse.SendCustomResponse(http.StatusOK, bookings)
}
