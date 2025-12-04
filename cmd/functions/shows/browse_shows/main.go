package main

import (
	"context"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	showrepository "eventro_aws/internals/repository/show_repository"
	showservice "eventro_aws/internals/services/show_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var showService showservice.ShowServiceI

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	showRepo := showrepository.NewShowRepositoryDDB(ddb, "eventro")
	showService = showservice.NewShowService(showRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(BrowseShows)))
}

func BrowseShows(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	city := event.QueryStringParameters["city"]
	eventID := event.QueryStringParameters["eventID"]
	date := event.QueryStringParameters["date"]
	venueID := event.QueryStringParameters["venueID"]

	role, _ := authenticationmiddleware.GetUserRole(ctx)
	var hostID string
	if strings.ToLower(role) == "host" {
		hostID, _ = authenticationmiddleware.GetUserEmail(ctx)
	}

	shows, err := showService.BrowseShows(ctx, eventID, city, date, venueID, hostID)
	if err != nil {
		return customresponse.LambdaError(http.StatusInternalServerError, err.Error())
	}

	return customresponse.SendCustomResponse(http.StatusOK, shows)
}
