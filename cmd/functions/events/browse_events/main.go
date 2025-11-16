package main

import (
	"context"
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
	isBlockedParam := event.QueryStringParameters["isBlocked"]

	var isBlocked *bool
	if isBlockedParam != "" {
		role, err := authenticationmiddleware.GetUserRole(ctx)
		if err != nil || role != "Admin" {
			return customresponse.SendCustomResponse(403, "only admin can access blocked events")
		}
		val := isBlockedParam == "true"
		isBlocked = &val
	}

	filter := models.EventFilter{
		EventID:    event.QueryStringParameters("eventID"),
		Name:       event.QueryStringParameters("eventname"),
		Category:   event.QueryStringParameters("category"),
		Location:   event.QueryStringParameters("location"),
		IsBlocked:  isBlocked,
		ArtistName: event.QueryStringParameters("artistName"),
	}

	events, err := eventService.BrowseEvents(ctx, filter)
	if err != nil {
		return customresponse.SendCustomResponse(500, "internal server error")
	}
	return customresponse.SendCustomResponse(200, events)

}
