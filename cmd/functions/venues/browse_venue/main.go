package main

import (
	"context"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	venuerepository "eventro_aws/internals/repository/venue_repository"
	venueservice "eventro_aws/internals/services/venue_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

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
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(BrowseVenues)))
}

func BrowseVenues(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	venueID := event.PathParameters["venueID"]

	_, err := authenticationmiddleware.GetUserRole(ctx)

	if err != nil {
		return customresponse.SendCustomResponse(
			http.StatusUnauthorized,
			"user not logged in",
		)
	}

	venue, err := venueService.GetVenueByID(ctx, venueID)
	if err != nil {
		return customresponse.SendCustomResponse(
			http.StatusInternalServerError,
			"Failed to fetch venues: "+err.Error(),
		)
	}

	return customresponse.SendCustomResponse(http.StatusOK, venue)
}
