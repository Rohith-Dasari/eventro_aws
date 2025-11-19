package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	artistrepository "eventro_aws/internals/repository/artist_repository"
	artistservice "eventro_aws/internals/services/artist_service"
	customresponse "eventro_aws/internals/utils"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var artistService artistservice.Artistservice

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	artistRepo := artistrepository.NewArtistRepositoryDDB(ddb, "eventro")
	artistService = artistservice.NewArtistService(artistRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(CreateArtist)))
}

type CreateArtistRequest struct {
	Name string `json:"name"`
	Bio  string `json:"bio"`
}

type CreateArtistResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Bio  string `json:"bio"`
}

func CreateArtist(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	role, err := authenticationmiddleware.GetUserRole(ctx)
	if err != nil || role != "Admin" {
		return customresponse.LambdaError(http.StatusForbidden, "forbidden: admin only")

	}

	var req CreateArtistRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.SendCustomResponse(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		return customresponse.LambdaError(http.StatusBadRequest, "artist name is required")
	}

	err = artistService.CreateArtist(ctx, req.Name, req.Bio)
	if err != nil {
		return customresponse.LambdaError(http.StatusBadRequest, err.Error())
	}

	return customresponse.SendCustomResponse(http.StatusOK, "successfully created artist")
}
