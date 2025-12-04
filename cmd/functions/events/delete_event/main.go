package main

import (
	"context"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	eventrepository "eventro_aws/internals/repository/event_repository"
	eventservice "eventro_aws/internals/services/event_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var eventService eventservice.EventServiceI

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	eventRepo := eventrepository.NewEventRepoDDB(ddb, "eventro")
	eventService = eventservice.NewEventService(eventRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(deleteEvent)))
}

func deleteEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil || strings.ToLower(role) != "admin" {
		return customresponse.LambdaError(403, "only admin can delete event")

	}

	eventID := event.QueryStringParameters["eventID"]
	if eventID == "" {
		return customresponse.LambdaError(400, "eventID is required in path")
	}

	err = eventService.DeleteEvent(ctx, eventID)
	if err != nil {
		return customresponse.LambdaError(500, "Failed to delete event")

	}

	return customresponse.SendCustomResponse(200, "successfully deleted", nil)
}
