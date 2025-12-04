package eventrepository

import (
	"context"
	"eventro_aws/internals/models"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type EventDDB struct {
	EventID     string   `dynamodbav:"pk"`
	SK          string   `dynamodbav:"sk"`
	EventName   string   `dynamodbav:"event_name"`
	Description string   `dynamodbav:"description"`
	Duration    string   `dynamodbav:"duration"`
	Category    string   `dynamodbav:"category"`
	IsBlocked   bool     `dynamodbav:"is_blocked"`
	ArtistIDs   []string `dynamodbav:"artist_ids"`
	ArtistNames []string `dynamodbav:"artist_names"`
}

type EventRepositoryDDB struct {
	db        *dynamodb.Client
	TableName string
}

func NewEventRepoDDB(db *dynamodb.Client, tableName string) *EventRepositoryDDB {
	return &EventRepositoryDDB{db: db, TableName: tableName}
}

func (er *EventRepositoryDDB) Create(ctx context.Context, event *models.Event) error {
	var artistNames []string

	for _, artistID := range event.ArtistIDs {
		artistPK := "ARTIST#" + artistID
		input := &dynamodb.QueryInput{
			TableName:              aws.String(er.TableName),
			KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: artistPK},
				":sk": &types.AttributeValueMemberS{Value: "NAME#"},
			},
		}

		result, err := er.db.Query(ctx, input)
		if err != nil {
			return err
		}

		if len(result.Items) > 0 {
			var artistData struct {
				ArtistName string `dynamodbav:"sk"`
			}
			err = attributevalue.UnmarshalMap(result.Items[0], &artistData)
			if err != nil {
				return err
			}
			artistNames = append(artistNames, strings.TrimPrefix(artistData.ArtistName, "NAME#"))
		}
	}

	dbItem := map[string]any{
		"pk":           "EVENT#" + event.ID,
		"sk":           "DETAILS",
		"event_name":   event.Name,
		"description":  event.Description,
		"duration":     event.Duration,
		"category":     string(event.Category),
		"is_blocked":   event.IsBlocked,
		"artist_ids":   event.ArtistIDs,
		"artist_names": artistNames,
	}

	itemAV, err := attributevalue.MarshalMap(dbItem)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = er.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(er.TableName),
		Item:      itemAV,
	})
	if err != nil {
		log.Printf("Couldn't put item into table %s: %v\n", er.TableName, err)
		return err
	}
	sk := fmt.Sprintf("EVENT_NAME#%s#EVENT_ID#%s", event.Name, event.ID)

	nameIndex := map[string]any{
		"pk": "EVENTS",
		"sk": sk,
	}

	nameAV, err := attributevalue.MarshalMap(nameIndex)
	if err != nil {
		return fmt.Errorf("marshal name-index error: %w", err)
	}

	_, err = er.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(er.TableName),
		Item:      nameAV,
	})
	if err != nil {
		return fmt.Errorf("put name-index error: %w", err)
	}

	return nil
}

func (er *EventRepositoryDDB) GetByID(ctx context.Context, eventID string) (*models.EventDTO, error) {
	var pk string

	if strings.HasPrefix(eventID, "EVENT#") {
		pk = eventID
	} else {
		pk = "EVENT#" + eventID
	}

	out, err := er.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(er.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("ddb get error: %w", err)
	}

	if len(out.Item) == 0 {
		return &models.EventDTO{}, nil
	}

	var eddb EventDDB
	if err := attributevalue.UnmarshalMap(out.Item, &eddb); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	dto := &models.EventDTO{
		EventID:     eventID,
		EventName:   eddb.EventName,
		Description: eddb.Description,
		Duration:    eddb.Duration,
		Category:    eddb.Category,
		IsBlocked:   eddb.IsBlocked,
		ArtistNames: eddb.ArtistNames,
		ArtistIDs:   eddb.ArtistIDs,
	}

	return dto, nil
}

func (er *EventRepositoryDDB) Update(ctx context.Context, eventID string, isBlocked bool) error {
	var pk string
	if strings.HasPrefix(eventID, "EVENT#") {
		pk = eventID
	} else {
		pk = "EVENT#" + eventID
	}
	_, err := er.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(er.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
		UpdateExpression: aws.String("SET is_blocked = :b"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":b": &types.AttributeValueMemberBOOL{Value: isBlocked},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	})

	if err != nil {
		return fmt.Errorf("failed to update is_blocked: %w", err)
	}

	return nil
}

func (er *EventRepositoryDDB) Delete(ctx context.Context, id string) error {
	key, err := attributevalue.MarshalMap(map[string]string{
		"pk": "EVENT#" + id,
		"sk": "DETAILS",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal key: %w", err)
	}
	_, err = er.db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(er.TableName),
		Key:       key,
	})
	if err != nil {
		log.Printf("delete failed: %v", err)
	}
	return err
}

func (er *EventRepositoryDDB) GetEventsByCity(ctx context.Context, city string) ([]*models.EventDTO, error) {
	pk := "CITY#" + city
	exprVals := map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: pk},
	}

	resp, err := er.db.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(er.TableName),
		KeyConditionExpression:    aws.String("pk = :pk"),
		ExpressionAttributeValues: exprVals,
	})
	if err != nil {
		log.Printf("Couldn't query events for city %s: %v\n", city, err)
		return nil, err
	}

	if len(resp.Items) == 0 {
		return []*models.EventDTO{}, nil
	}

	eventIDs := []string{}
	for _, item := range resp.Items {
		var dbRec struct {
			SK string `dynamodbav:"sk"`
		}
		if err := attributevalue.UnmarshalMap(item, &dbRec); err != nil {
			log.Printf("failed to unmarshal item for city %s: %v\n", city, err)
			continue
		}

		eventID := strings.TrimPrefix(dbRec.SK, "EVENT#")
		if eventID == "" {
			continue
		}
		eventIDs = append(eventIDs, eventID)

	}

	if len(eventIDs) == 0 {
		return []*models.EventDTO{}, nil
	}

	return er.BatchGetEvents(ctx, eventIDs)
}

func (er *EventRepositoryDDB) BatchGetEvents(ctx context.Context, eventIDs []string) ([]*models.EventDTO, error) {
	if len(eventIDs) == 0 {
		return []*models.EventDTO{}, nil
	}

	events := []*models.EventDTO{}

	for start := 0; start < len(eventIDs); start += 100 {
		end := start + 100
		if end > len(eventIDs) {
			end = len(eventIDs)
		}

		chunk := eventIDs[start:end]
		keys := make([]map[string]types.AttributeValue, 0, len(chunk))
		for _, id := range chunk {
			pk := id
			if !strings.HasPrefix(id, "EVENT#") {
				pk = "EVENT#" + id
			}
			keys = append(keys, map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{Value: pk},
				"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
			})
		}

		req := &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				er.TableName: {
					Keys: keys,
				},
			},
		}

		for {
			resp, err := er.db.BatchGetItem(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("batch get error: %w", err)
			}

			// Unmarshal found items
			for _, item := range resp.Responses[er.TableName] {
				var eddb EventDDB
				if err := attributevalue.UnmarshalMap(item, &eddb); err != nil {
					continue
				}

				eventID := strings.TrimPrefix(eddb.EventID, "EVENT#")

				events = append(events, &models.EventDTO{
					EventID:     eventID,
					EventName:   eddb.EventName,
					Description: eddb.Description,
					Duration:    eddb.Duration,
					Category:    eddb.Category,
					IsBlocked:   eddb.IsBlocked,
					ArtistIDs:   eddb.ArtistIDs,
					ArtistNames: eddb.ArtistNames,
				})
			}

			unprocessed := resp.UnprocessedKeys
			u := unprocessed[er.TableName]

			if len(u.Keys) == 0 {
				break
			}

			req.RequestItems = map[string]types.KeysAndAttributes{
				er.TableName: {
					Keys: u.Keys,
				},
			}
		}
	}

	return events, nil
}

func (er *EventRepositoryDDB) GetEventsHostedByHost(ctx context.Context, hostID string) ([]*models.EventDTO, error) {
	pk := "HOST#" + hostID
	skPrefix := "EVENT#"

	out, err := er.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(er.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
			":sk": &types.AttributeValueMemberS{Value: skPrefix},
		},
	})
	if err != nil {
		return nil, err
	}

	eventIDs := []string{}
	for _, item := range out.Items {
		var row struct {
			SK string `dynamodbav:"sk"`
		}
		if err := attributevalue.UnmarshalMap(item, &row); err != nil {
			return nil, err
		}

		eventIDs = append(eventIDs, strings.TrimPrefix(row.SK, "EVENT#"))

	}
	return er.BatchGetEvents(ctx, eventIDs)
}

func (er *EventRepositoryDDB) GetEventsByName(ctx context.Context, name string) ([]*models.EventDTO, error) {
	ids, err := er.SearchByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error from SearchByName: %s", err.Error())
	}
	if len(ids) == 0 {
		return []*models.EventDTO{}, nil
	}
	return er.BatchGetEvents(ctx, ids)
}
func (er *EventRepositoryDDB) GetBlockedEvents(ctx context.Context) ([]*models.EventDTO, error) {
	ids, err := er.SearchByName(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("error from SearchByName: %s", err.Error())
	}
	if len(ids) == 0 {
		return []*models.EventDTO{}, nil
	}

	var eventsDTO []*models.EventDTO
	events, _ := er.BatchGetEvents(ctx, ids)

	for _, event := range events {
		if event.IsBlocked {
			eventsDTO = append(eventsDTO, event)
		}
	}
	return eventsDTO, nil

}

func (er *EventRepositoryDDB) SearchByName(ctx context.Context, namePrefix string) ([]string, error) {

	skPrefix := "EVENT_NAME#" + namePrefix

	out, err := er.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(er.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: "EVENTS"},
			":skPrefix": &types.AttributeValueMemberS{Value: skPrefix},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error from Query: %s", err.Error())
	}

	eventIDs := []string{}
	for _, item := range out.Items {
		attr, ok := item["sk"].(*types.AttributeValueMemberS)
		if !ok {
			continue
		}
		sk := attr.Value

		parts := strings.SplitN(sk, "EVENT_ID#", 2)
		if len(parts) != 2 {
			continue
		}

		eventID := parts[1]
		eventIDs = append(eventIDs, eventID)
	}

	return eventIDs, nil
}
