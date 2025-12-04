package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	showrepository "eventro_aws/internals/repository/show_repository"
	showservice "eventro_aws/internals/services/show_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type UpdateShowRequest struct {
	IsBlocked bool `json:"is_blocked,omitempty"`
}

var showService showservice.ShowServiceI

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	showRepo := showrepository.NewShowRepositoryDDB(ddb, "eventro")
	showService = showservice.NewShowService(showRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(UpdateShow)))
}

func UpdateShow(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	showID := event.PathParameters["showID"]

	userID, err := authenticationmiddleware.GetUserID(ctx)
	if err != nil || userID == "" {
		return customresponse.LambdaError(http.StatusUnauthorized, "user unauthorised")
	}

	var req UpdateShowRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.SendCustomResponse(http.StatusBadRequest, "invalid request body")
	}

	err = showService.UpdateShow(ctx, showID, userID, req.IsBlocked)
	if err != nil {
		return customresponse.LambdaError(http.StatusInternalServerError, "Failed to update show: "+err.Error())
	}

	return customresponse.SendCustomResponse(http.StatusOK, "successfully updated")
}
