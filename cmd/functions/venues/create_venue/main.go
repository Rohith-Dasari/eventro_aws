package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	venuerepository "eventro_aws/internals/repository/venue_repository"
	venueservice "eventro_aws/internals/services/venue_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type CreateVenueRequest struct {
	Name                 string `json:"name"`
	City                 string `json:"city"`
	State                string `json:"state"`
	IsSeatLayoutRequired bool   `json:"is_seat_layout_required"`
}

var venueService venueservice.VenueServiceI

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	venueRepo := venuerepository.NewVenueRepositoryDDB(ddb, "eventro")
	venueService = venueservice.NewVenueService(venueRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(CreateVenue)))
}

func CreateVenue(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil {
		return customresponse.SendCustomResponse(http.StatusForbidden, "unable to get user role")
	}
	if strings.ToLower(role) != "host" {
		return customresponse.SendCustomResponse(http.StatusForbidden, "only host can create venue")
	}
	hostID, err := authenticationmiddleware.GetUserEmail(ctx)
	if err != nil || hostID == "" {
		return customresponse.LambdaError(http.StatusUnauthorized, "not authorised")
	}

	var req CreateVenueRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.SendCustomResponse(400, "invalid request body")
	}

	venue, err := venueService.CreateVenue(
		ctx,
		hostID,
		req.Name,
		req.City,
		req.State,
		req.IsSeatLayoutRequired,
	)
	if err != nil {
		return customresponse.SendCustomResponse(500, err.Error())
	}

	return customresponse.SendCustomResponse(http.StatusCreated, venue)
}
