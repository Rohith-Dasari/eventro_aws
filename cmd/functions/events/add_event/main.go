package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	"eventro_aws/internals/models"
	eventrepository "eventro_aws/internals/repository/event_repository"
	eventservice "eventro_aws/internals/services/event_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var eventService eventservice.EventService

type CreateEventRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Category    string   `json:"category"`
	ArtistIDs   []string `json:"artists_ids"`
	ArtistNames []string `json:"artist_names,omitempty"`
}

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	eventRepo := eventrepository.NewEventRepoDDB(ddb, "eventro")
	eventService = eventservice.NewEventService(eventRepo)
}

func main() {
	lambda.Start(CreateEvent)
}

func CreateEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil || (role != "Admin" && role != "Host") {
		return customresponse.LambdaError(403, "Only admin and host authorised")
	}

	var req CreateEventRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.LambdaError(400, "invalid request body")
	}

	createdEvent, nil := eventService.CreateNewEvent(ctx, req.Name, req.Description, req.Duration, models.EventCategory(req.Category), req.ArtistIDs)

	if err != nil {
		return customresponse.LambdaError(500, "Internal server error")
	}

	body, _ := json.Marshal(createdEvent)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil

}
