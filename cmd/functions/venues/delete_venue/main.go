package main

import (
	"context"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	venuerepository "eventro_aws/internals/repository/venue_repository"
	venueservice "eventro_aws/internals/services/venue_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

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
	lambda.Start(DeleteVenue)
}

func DeleteVenue(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	venueID := event.PathParameters["venueID"]
	if venueID == "" {
		return customresponse.LambdaError(http.StatusBadRequest, "missing venueID in path param")
	}

	userID, err := authenticationmiddleware.GetUserID(ctx)
	if err != nil || userID == "" {
		return customresponse.LambdaError(http.StatusUnauthorized, "user unauthorised")
	}

	userRole, err := authenticationmiddleware.GetUserRole(ctx)
	userRole = strings.ToLower(userRole)
	if err != nil || (userRole != "host" && userRole != "admin") {
		return customresponse.LambdaError(http.StatusForbidden, "only admin and host can delete venue")
	}

	if err := venueService.DeleteVenue(ctx, venueID, userID, userRole); err != nil {
		return customresponse.LambdaError(http.StatusInternalServerError, err.Error())
	}

	return customresponse.SendCustomResponse(http.StatusOK, "Successfully deleted")
}
