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
	lambda.Start(Signup)
}

func Signup(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	var req models.SignupRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		body, _ := json.Marshal(map[string]string{"message": "invalid request body"})
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	if req.Username == "" || req.Email == "" || req.PhoneNumber == "" || req.Password == "" {
		body, _ := json.Marshal(map[string]string{"message": "invalid request body"})
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	if len(req.Password) < 12 {
		body, _ := json.Marshal(map[string]string{"message": "Password should be of atleast 12 alphanumeric characters and a symbol"})
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       string(body),
		}, nil
	}

	user, err := authService.Signup(ctx, req.Username, req.Email, req.PhoneNumber, req.Password)
	if err != nil {
		message := err.Error()
		body, _ := json.Marshal(map[string]string{"message": message})
		return events.APIGatewayProxyResponse{
			StatusCode: 409,
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

	return customresponse.SendCustomResponse(200, res)
}
