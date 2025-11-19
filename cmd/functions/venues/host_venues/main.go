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
	lambda.Start(authenticationmiddleware.AuthorizedInvoke(GetHostVenues))
}

func GetHostVenues(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	hostID := event.PathParameters["hostId"]

	userRole, err := authenticationmiddleware.GetUserRole(ctx)
	userRole = strings.ToLower(userRole)

	if err != nil || (userRole != "host" && userRole != "admin") {
		return customresponse.SendCustomResponse(
			http.StatusUnauthorized,
			"user unauthorised",
		)
	}
	fmt.Println("hostID: " + hostID)
	hostEmail, _ := authenticationmiddleware.GetUserEmail(ctx)

	venues, err := venueService.GetHostVenues(ctx, hostEmail)
	if err != nil {
		return customresponse.SendCustomResponse(
			http.StatusInternalServerError,
			"Failed to fetch venues: "+err.Error(),
		)
	}

	return customresponse.SendCustomResponse(http.StatusOK, venues)
}
