package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
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
	lambda.Start(BrowseEvents)
}

func BrowseEvents(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// isBlockedParam := event.QueryStringParameters["isBlocked"]
	cityParam := event.QueryStringParameters["city"]

	resEvents, err := eventService.BrowseEvents(ctx, cityParam)
	if err != nil {
		return customresponse.LambdaError(500, "internal server error")
	}

	body, err := json.Marshal(resEvents)
	if err != nil {
		return customresponse.LambdaError(500, "failed to marshal events")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil

}
