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

type UpdateVenueRequest struct {
	IsBlocked bool `json:"is_blocked,omitempty"`
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
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(UpdateVenue)))
}

func UpdateVenue(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	venueID := event.PathParameters["venueID"]
	if venueID == "" {
		return customresponse.LambdaError(400, "invalid request")
	}

	userID, err := authenticationmiddleware.GetUserEmail(ctx)
	if err != nil || userID == "" {
		return customresponse.LambdaError(403, "not authorised")
	}
	userRole, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil {
		return customresponse.LambdaError(403, "unable to get user role")
	}
	if strings.ToLower(userRole) != "host" && strings.ToLower(userRole) != "admin" {
		return customresponse.LambdaError(403, "only admin and host is authorised")
	}

	var req UpdateVenueRequest

	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.LambdaError(400, "invalid request body")
	}

	err = venueService.UpdateVenue(ctx, venueID, userID, userRole, req.IsBlocked)
	if err != nil {
		return customresponse.LambdaError(http.StatusInternalServerError, err.Error())
	}
	return customresponse.SendCustomResponse(http.StatusOK, "successfully moderated", nil)
}
