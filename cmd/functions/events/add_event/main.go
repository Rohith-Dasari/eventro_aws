package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	"eventro_aws/internals/models"
	eventrepository "eventro_aws/internals/repository/event_repository"
	eventservice "eventro_aws/internals/services/event_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var eventService eventservice.EventServiceI

type CreateEventRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Category    string   `json:"category"`
	ArtistIDs   []string `json:"artist_ids"`
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
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(CreateEvent)))
}

func CreateEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil {
		return customresponse.LambdaError(http.StatusUnauthorized, err.Error())
	}
	if strings.ToLower(role) != "admin" && strings.ToLower(role) != "host" {
		return customresponse.LambdaError(403, "Only admin and host authorised")
	}

	var req CreateEventRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.LambdaError(400, "invalid request body")
	}

	createdEvent, err := eventService.CreateNewEvent(ctx, req.Name, req.Description, req.Duration, models.EventCategory(req.Category), req.ArtistIDs)

	if err != nil {
		return customresponse.LambdaError(500, err.Error())
	}

	return customresponse.SendCustomResponse(200, "successfully created", createdEvent)

}

// type body struct {
// 	statusCode int
// 	data       any
// }
