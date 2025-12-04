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
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type CreateShowRequest struct {
	EventID  string  `json:"event_id"`
	VenueID  string  `json:"venue_id"`
	Price    float64 `json:"price"`
	ShowDate string  `json:"show_date"`
	ShowTime string  `json:"show_time"`
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
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(CreateShow)))
}

func CreateShow(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID, err := authenticationmiddleware.GetUserEmail(ctx)
	if err != nil || userID == "" {

		return customresponse.LambdaError(http.StatusUnauthorized, "not authorised")
	}

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil {
		return customresponse.LambdaError(403, "unable to get role")
	}
	if strings.ToLower(role) != "host" {
		return customresponse.LambdaError(403, "Only admin and host authorised")
	}

	var req CreateShowRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.LambdaError(http.StatusBadRequest, "invalid request body")

	}

	parsedDate, err := time.Parse("2006-01-02", req.ShowDate)
	if err != nil {
		return customresponse.LambdaError(http.StatusBadRequest, "Invalid date format, expected YYYY-MM-DD")
	}

	err = showService.CreateShow(
		ctx,
		req.EventID,
		req.VenueID,
		userID,
		req.Price,
		parsedDate,
		req.ShowTime,
	)
	if err != nil {
		return customresponse.LambdaError(http.StatusInternalServerError, err.Error())
	}

	return customresponse.SendCustomResponse(http.StatusOK, "successfully created", nil)
}
