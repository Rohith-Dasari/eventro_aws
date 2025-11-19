package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
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
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(CreateBooking)))
}

type CreateBookingRequest struct {
	UserID string   `json:"user_id,omitempty"`
	ShowID string   `json:"show_id"`
	Seats  []string `json:"seats"`
}

func CreateBooking(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil || (role != "Customer" && role != "Admin") {
		return customresponse.SendCustomResponse(http.StatusForbidden, "only customers or admin for customers make booking")
	}
	authUserID, err := authenticationmiddleware.GetUserEmail(ctx)
	if err != nil || authUserID == "" {
		return customresponse.SendCustomResponse(401, err.Error())
	}

	var req CreateBookingRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.SendCustomResponse(400, "invalid request body")
	}

	if req.ShowID == "" || len(req.Seats) == 0 {
		return customresponse.SendCustomResponse(400, "invalid request")
	}
	userID := authUserID
	if role == "Admin" && req.UserID != "" {
		userID = req.UserID
	}

	booking, err := bookingService.AddBooking(ctx, userID, req.ShowID, req.Seats)
	if err != nil {
		return customresponse.LambdaError(http.StatusBadRequest, err.Error())
	}

	return customresponse.SendCustomResponse(http.StatusOK, booking)

}
