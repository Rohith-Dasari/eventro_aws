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
	"net/http"

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
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(GetEventByID)))
}

func GetEventByID(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	eventID := event.PathParameters["eventID"]

	resEvent, err := eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return customresponse.LambdaError(500, "internal server error: "+err.Error())
	}

	body, err := json.Marshal(resEvent)
	if err != nil {
		return customresponse.LambdaError(500, "failed to marshal events")
	}

	return customresponse.SendCustomResponse(http.StatusOK, "successfully retrived event", body)
}
