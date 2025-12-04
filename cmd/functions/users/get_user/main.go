package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	authenticationmiddleware "eventro_aws/internals/middleware/authentication_middleware"
	corsmiddleware "eventro_aws/internals/middleware/cors_middleware"
	userrepository "eventro_aws/internals/repository/user_repository"
	"eventro_aws/internals/services/userservice"
	customresponse "eventro_aws/internals/utils"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var userService userservice.UserServiceI

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	userRepo := userrepository.NewUserRepoDDB(ddb, "eventro")
	userService = userservice.NewUserService(userRepo)
}

func main() {
	lambda.Start(corsmiddleware.WithCORS(authenticationmiddleware.AuthorizedInvoke(GetUserByID)))
}

func GetUserByID(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	emailID := event.PathParameters["emailID"]

	resEvent, err := userService.GetUserByMailID(ctx, emailID)
	if err != nil {
		return customresponse.LambdaError(500, "internal server error: "+err.Error())
	}

	body, err := json.Marshal(resEvent)
	if err != nil {
		return customresponse.LambdaError(500, "failed to marshal events")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}
