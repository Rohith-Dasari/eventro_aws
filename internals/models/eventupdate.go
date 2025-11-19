package models

type EventUpdate struct {
	IsBlocked bool `json:"is_blocked" dynamodbav:"is_blocked"`
}
