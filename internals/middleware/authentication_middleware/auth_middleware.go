package authenticationmiddleware

import (
	"context"
	"errors"
	"eventro_aws/internals/services/authorisation"
	customresponse "eventro_aws/internals/utils"

	"github.com/aws/aws-lambda-go/events"
)

type contextKey string

const (
	ContextUserIDKey    contextKey = "userID"
	ContextUserEmailKey contextKey = "userEmail"
	ContextUserRoleKey  contextKey = "userRole"
)

func AuthorizedInvoke(fn func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		authHeader := req.Headers["Authorization"]
		if authHeader == "" {
			return customresponse.LambdaError(401, "Unauthorized")
		}

		tokenString := authHeader[len("Bearer "):]

		claims, err := authorisation.ValidateJWT(tokenString)
		if err != nil {
			return customresponse.LambdaError(401, "Unauthorized: "+err.Error())
		}
		authCtx := context.WithValue(ctx, ContextUserIDKey, claims.UserID)
		authCtx = context.WithValue(authCtx, ContextUserEmailKey, claims.Email)
		authCtx = context.WithValue(authCtx, ContextUserRoleKey, claims.Role)

		return fn(authCtx, req)
	}
}

func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(ContextUserIDKey).(string)
	if !ok {
		return "", errors.New("userID not found in context")
	}
	return userID, nil
}

func GetUserEmail(ctx context.Context) (string, error) {
	email, ok := ctx.Value(ContextUserEmailKey).(string)
	if !ok {
		return "", errors.New("userEmail not found in context")
	}
	return email, nil
}

func GetUserRole(ctx context.Context) (string, error) {
	role, ok := ctx.Value(ContextUserRoleKey).(string)
	if !ok {
		return "", errors.New("userRole not found in context")
	}
	return role, nil
}
