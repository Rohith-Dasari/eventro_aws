package eventrepository

import (
	"context"
	"errors"
	"eventro_aws/internals/models"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type EventRepositoryDDB struct {
	db        *dynamodb.Client
	TableName string
}

func NewEventRepoDDB(db *dynamodb.Client, tableName string) *EventRepositoryDDB {
	return &EventRepositoryDDB{db: db, TableName: tableName}
}

func (er *EventRepositoryDDB) Create(event *models.Event) error {
	ctx := context.Background()
	params, err := attributevalue.MarshalList([]interface{}{event.ID, "DETAILS", event.Name, event.Description, event.Duration, event.Category, event.IsBlocked})
	if err != nil {
		panic(err)
	}
	_, err = er.db.ExecuteStatement(ctx, &dynamodb.ExecuteStatementInput{
		Statement: aws.String(
			fmt.Sprintf("INSERT INTO \"%v\" VALUE {'pk':?,'sk':?,'event_name':?,'description':?,'duration'}", er.TableName)),
		Parameters: params,
	})
	if err != nil {
		log.Printf("Couldn't insert an item with PartiQL. Here's why: %v\n", err)
	}
	return err
}

func (er *EventRepositoryDDB) GetByID(id string) (*models.Event, error) {
	ctx := context.Background()
	event := &models.Event{}
	params, err := attributevalue.MarshalList([]interface{}{"EVENT#" + id, "DETAILS"})
	if err != nil {
		panic(err)
	}
	response, err := er.db.ExecuteStatement(ctx, &dynamodb.ExecuteStatementInput{
		Statement: aws.String(
			fmt.Sprintf("SELECT * FROM \"%v\" WHERE pk=? AND sk=?",
				er.TableName)),
		Parameters: params,
	})
	if len(response.Items) == 0 {
		return nil, errors.New("no id found")
	}
	if err != nil {
		log.Printf("Couldn't get info about %v. Here's why: %v\n", id, err)
	} else {
		err = attributevalue.UnmarshalMap(response.Items[0], &event)
		if err != nil {
			log.Printf("Couldn't unmarshal response. Here's why: %v\n", err)
		}
	}
	event.ID = strings.TrimPrefix(event.ID, "EVENT#")
	return event, nil
}

func (er *EventRepositoryDDB) List() ([]models.Event, error) {
	return []models.Event{}, nil
}

func (er *EventRepositoryDDB) Update(event *models.Event) error {
	return nil
}

func (er *EventRepositoryDDB) Delete(id string) error {
	return nil
}

func (er *EventRepositoryDDB) AddEventArtist(ea *models.EventArtist) error {
	return nil
}

func (er *EventRepositoryDDB) GetArtistsByEventID(eventID string) ([]models.Artist, error) {
	return []models.Artist{}, nil
}
func (er *EventRepositoryDDB) GetEventsByCity(city string) ([]models.Event, error) {
	return []models.Event{}, nil

}

func (er *EventRepositoryDDB) GetFilteredEvents(filter models.EventFilter) ([]models.EventResponse, error) {
	return []models.EventResponse{}, nil

}

func (er *EventRepositoryDDB) GetEventsHostedByHost(hostID string) ([]models.EventResponse, error) {
	return []models.EventResponse{}, nil
}
