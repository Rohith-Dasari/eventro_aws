package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	"eventro_aws/internals/models"
	userrepository "eventro_aws/internals/repository/user_repository"
	"eventro_aws/internals/services/authorisation"
	customresponse "eventro_aws/internals/utils"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var authService authorisation.AuthService

func init() {
	ddb, err := db.InitDB()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %v", err))
	}

	userRepo := userrepository.NewUserRepoDDB(ddb, "eventro")
	authService = *authorisation.NewAuthService(userRepo)
}

func main() {
	lambda.Start(Login)
}

func Login(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var req models.LoginRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return customresponse.LambdaError(400, "invalid request body ")
	}
	user, err := authService.ValidateLogin(ctx, req.Email, req.Password)

	if err != nil {
		message := err.Error()
		return customresponse.LambdaError(401, message)
	}

	token, err := authorisation.GenerateJWT(user.UserID, user.Email, string(user.Role))
	if err != nil {
		return customresponse.LambdaError(500, "failed to generate token")
	}

	res := models.LoginResponse{
		Token: token,
	}

	body, _ := json.Marshal(res)
	return customresponse.SendCustomResponse(200, body)
}
