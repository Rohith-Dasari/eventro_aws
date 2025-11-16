package main

import (
	"context"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	eventrepository "eventro_aws/internals/repository/event_repository"
	eventservice "eventro_aws/internals/services/event_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"

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
	lambda.Start(deleteEvent)
}

func deleteEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil || role != "Admin" {
		return customresponse.SendCustomResponse(403, "only admin can delete event")

	}

	eventID := r.PathValue("eventID")
	if eventID == "" {
		return customresponse.SendCustomResponse(400,"eventID is required in path")
	}

	err = h.EventService.DeleteEvent(r.Context(), eventID)
	if err != nil {
		responses.InternalServerError(w, "Failed to delete event")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
