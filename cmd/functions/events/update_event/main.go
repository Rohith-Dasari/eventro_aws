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

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	eventRepo := eventrepository.NewEventRepoDDB(ddb, "eventro")
	eventService = eventservice.NewEventService(eventRepo)
}

func main() {
	lambda.Start(authenticationmiddleware.AuthorizedInvoke(UpdateEvent))
}

func UpdateEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil || role != "Admin" {
		return customresponse.LambdaError(403, "only admin is authorised")
	}

	eventID := event.PathParameters["eventID"]

	if eventID == "" {
		return customresponse.LambdaError(400, "eventID is required")
	}

	var updateData models.EventUpdate
	if err := json.Unmarshal([]byte(event.Body), &updateData); err != nil {
		return customresponse.LambdaError(400, "invalid request body")
	}

	err = eventService.UpdateEvent(ctx, eventID, updateData.IsBlocked)
	if err != nil {
		return customresponse.LambdaError(500, "internal server error: "+err.Error())
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "successfully moderated event",
	}, nil
}
