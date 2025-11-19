package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	eventrepository "eventro_aws/internals/repository/event_repository"
	eventservice "eventro_aws/internals/services/event_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var eventService eventservice.EventService

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	eventRepo := eventrepository.NewEventRepoDDB(ddb, "eventro")
	eventService = eventservice.NewEventService(eventRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(EventsOfHost)))
}

func EventsOfHost(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	hostID := event.PathParameters["hostID"]
	if hostID == "" {
		hostID = event.QueryStringParameters["hostID"]
	}
	if hostID == "" {
		return customresponse.SendCustomResponse(400, "hostID is required")
	}

	hostEmail, err := authenticationmiddleware.GetUserEmail(ctx)
	if err != nil {
		return customresponse.LambdaError(403, "unable to get user email")
	}

	hostEvents, err := eventService.GetHostEvents(ctx, hostEmail)
	if err != nil {
		return customresponse.SendCustomResponse(500, "Failed to fetch events")
	}

	body, err := json.Marshal(hostEvents)
	if err != nil {
		return customresponse.SendCustomResponse(500, "failed to marshal response")
	}
	return customresponse.SendCustomResponse(200, body)
}
