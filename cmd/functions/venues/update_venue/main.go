package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	"eventro_aws/internals/models"
	venuerepository "eventro_aws/internals/repository/venue_repository"
	venueservice "eventro_aws/internals/services/venue_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type UpdateVenueRequest struct {
	Name                 *string `json:"name,omitempty"`
	City                 *string `json:"city,omitempty"`
	State                *string `json:"state,omitempty"`
	IsSeatLayoutRequired *bool   `json:"is_seat_layout_required,omitempty"`
	IsBlocked            *bool   `json:"is_blocked,omitempty"`
}

var venueService venueservice.VenueService

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	venueRepo := venuerepository.NewVenueRepositoryDDB(ddb, "eventro")
	venueService = venueservice.NewVenueService(venueRepo)
}

func main() {
	lambda.Start(authenticationmiddleware.AuthorizedInvoke(UpdateVenue))
}

func UpdateVenue(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	venueID := event.PathParameters["venueID"]
	if venueID == "" {
		return customresponse.SendCustomResponse(400, "invalid request")
	}

	userID, err := authenticationmiddleware.GetUserID(ctx)
	if err != nil || userID == "" {
		return customresponse.SendCustomResponse(403, "not authorised")
	}
	userRole, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil || userRole != "Host" {
		return customresponse.SendCustomResponse(403, "only host is authorised")
	}

	var req UpdateVenueRequest

	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.SendCustomResponse(400, "invalid request body")
	}
	fmt.Println("decoded")

	update := models.UpdateVenueData{
		Name:                 req.Name,
		City:                 req.City,
		State:                req.State,
		IsSeatLayoutRequired: req.IsSeatLayoutRequired,
		IsBlocked:            req.IsBlocked,
	}

	updatedVenue, err := venueService.UpdateVenue(ctx, venueID, userID, userRole, update)
	if err != nil {
		return customresponse.SendCustomResponse(http.StatusInternalServerError, err.Error())
	}
	return customresponse.SendCustomResponse(http.StatusOK, updatedVenue)
}
