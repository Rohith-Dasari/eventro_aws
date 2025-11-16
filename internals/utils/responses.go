package customresponse

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

type CustomResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func SendCustomResponse(statusCode int, data string) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(data)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil

}

func LambdaError(code int, msg string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]string{
		"message": msg,
	})
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       string(body),
	}
}
