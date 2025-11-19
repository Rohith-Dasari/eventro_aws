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

	eventRepo := eventrepository.NewEventRepoDDB(ddb, "eventro")
	eventService = eventservice.NewEventService(eventRepo)
}

func main() {
	lambda.Start(authenticationmiddleware.AuthorizedInvoke(BrowseEvents))
}

func BrowseEvents(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cityParam := event.QueryStringParameters["city"]
	nameParam := event.QueryStringParameters["name"]

	resEvents, err := eventService.BrowseEvents(ctx, cityParam, nameParam)
	if err != nil {
		return customresponse.LambdaError(500, "internal server error: "+err.Error())
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
