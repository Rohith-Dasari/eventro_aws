package main

import (
	"context"
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

var artistService artistservice.ArtistServiceI

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	artistRepo := artistrepository.NewArtistRepositoryDDB(ddb, "eventro")
	artistService = artistservice.NewArtistService(artistRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(BrowseArtists)))
}

func BrowseArtists(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	artistID := event.PathParameters["artistID"]
	if artistID != "" {
		artist, err := artistService.GetArtistByID(ctx, artistID)
		if err != nil {
			return customresponse.LambdaError(http.StatusInternalServerError, err.Error())

		}
		return customresponse.SendCustomResponse(http.StatusOK, artist)
	} else {
		return customresponse.LambdaError(http.StatusBadRequest, "missing artistID")
	}
}
