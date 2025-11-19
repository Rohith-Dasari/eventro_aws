package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
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

	eventRepo := eventrepository.NewEventRepoDDB(ddb, "events")
	eventService = eventservice.NewEventService(eventRepo)
}

func main() {
	lambda.Start(authenticationmiddleware.AuthorizedInvoke(EventsOfHost))
}

func EventsOfHost(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get hostID from path parameters first, then query string
	hostID := event.PathParameters["hostID"]
	if hostID == "" {
		hostID = event.QueryStringParameters["hostID"]
	}
	if hostID == "" {
		return customresponse.SendCustomResponse(400, "hostID is required")
	}

	hostEvents, err := eventService.GetHostEvents(ctx, hostID)
	if err != nil {
		return customresponse.SendCustomResponse(500, "Failed to fetch events")
	}

	body, err := json.Marshal(hostEvents)
	if err != nil {
		return customresponse.SendCustomResponse(500, "failed to marshal response")
	}
	return customresponse.SendCustomResponse(200, body)
}
