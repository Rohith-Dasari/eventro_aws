package customresponse

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

type CustomResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func SendCustomResponse(statusCode int, data any) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(data)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil

}

func LambdaError(code int, msg string) (events.APIGatewayProxyResponse, error) {
	body, err := json.Marshal(map[string]string{
		"message": msg,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "couldn't marshal response",
		}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       string(body),
	}, nil
}
