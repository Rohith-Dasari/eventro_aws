package main

import (
	"context"
	"encoding/json"
	"eventro_aws/db"
	"eventro_aws/internals/models"
	userrepository "eventro_aws/internals/repository/user_repository"
	"eventro_aws/internals/services/authorisation"
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
		body, _ := json.Marshal(map[string]string{"message": "invalid request body"})
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	user, err := authService.ValidateLogin(ctx, req.Email, req.Password)
	message := err.Error()
	if err != nil {
		body, _ := json.Marshal(map[string]string{"message": message})
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       string(body),
		}, nil
	}

	token, err := authorisation.GenerateJWT(user.UserID, user.Email, string(user.Role))
	if err != nil {
		body, _ := json.Marshal(map[string]string{"message": "failed to generate token"})
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       string(body),
		}, nil
	}

	res := models.LoginResponse{
		Token: token,
	}

	body, _ := json.Marshal(res)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}
