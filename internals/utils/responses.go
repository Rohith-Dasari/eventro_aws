package customresponse

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

type CustomResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Data       any    `json:"data,omitempty"`
}

func SendCustomResponse(statusCode int, message string, data any) (events.APIGatewayProxyResponse, error) {
	cr := CustomResponse{StatusCode: statusCode, Message: message, Data: data}
	body, _ := json.Marshal(cr)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil

}

func LambdaError(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	cr := CustomResponse{StatusCode: statusCode, Message: message}
	body, _ := json.Marshal(cr)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil

}
