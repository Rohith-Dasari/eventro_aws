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
}

type EventRepositoryDDB struct {
	db        *dynamodb.Client
	TableName string
}

func NewEventRepoDDB(db *dynamodb.Client, tableName string) *EventRepositoryDDB {
	return &EventRepositoryDDB{db: db, TableName: tableName}
}

func (er *EventRepositoryDDB) Create(ctx context.Context, event *models.Event) error {
	dbItem := map[string]any{
		"pk":          "EVENT#" + event.ID,
		"sk":          "DETAILS",
		"event_name":  event.Name,
		"description": event.Description,
		"duration":    event.Duration,
		"category":    string(event.Category),
		"is_blocked":  event.IsBlocked,
		"artist_ids":  event.ArtistIDs,
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

	if out.Item == nil || len(out.Item) == 0 {
		return &models.EventDTO{}, nil
	}

	var eddb EventDDB
	if err := attributevalue.UnmarshalMap(out.Item, &eddb); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	var artistNames []string
	var artistBios []string

	for _, artistID := range eddb.ArtistIDs {

		artistPK := "ARTIST#" + artistID

		artistOut, err := er.db.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(er.TableName),
			KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: artistPK},
				":skPrefix": &types.AttributeValueMemberS{Value: "NAME#"},
			},
			Limit: aws.Int32(1),
		})
		if err != nil || len(artistOut.Items) == 0 {
			continue
		}

		var artist models.ArtistDTO

		if err := attributevalue.UnmarshalMap(artistOut.Items[0], &artist); err != nil {
			continue
		}
		artist.Name = strings.TrimPrefix(artist.Name, "NAME#")

		artistNames = append(artistNames, artist.Name)
		artistBios = append(artistBios, artist.Bio)
	}

	dto := &models.EventDTO{
		EventID:     eventID,
		EventName:   eddb.EventName,
		Description: eddb.Description,
		Duration:    eddb.Duration,
		Category:    eddb.Category,
		IsBlocked:   eddb.IsBlocked,
		ArtistNames: artistNames,
		ArtistBios:  artistBios,
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
	// ctx := context.Background()
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

	var results []*models.EventDTO
	for _, item := range resp.Items {
		var dbRec struct {
			PK          string   `dynamodbav:"pk"`
			SK          string   `dynamodbav:"sk"`
			EventName   string   `dynamodbav:"event_name"`
			Description string   `dynamodbav:"description"`
			Duration    string   `dynamodbav:"duration"`
			Category    string   `dynamodbav:"category"`
			IsBlocked   bool     `dynamodbav:"is_blocked"`
			ArtistIDs   []string `dynamodbav:"artist_ids"`
		}
		if err := attributevalue.UnmarshalMap(item, &dbRec); err != nil {
			log.Printf("failed to unmarshal item for city %s: %v\n", city, err)
			continue
		}

		eventID := strings.TrimPrefix(dbRec.SK, "EVENT#")

		var artistNames []string
		var artistBios []string
		for _, artistID := range dbRec.ArtistIDs {
			artistPK := "ARTIST#" + artistID
			artistResp, aerr := er.db.Query(ctx, &dynamodb.QueryInput{
				TableName:              aws.String(er.TableName),
				KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :skPrefix)"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pk":       &types.AttributeValueMemberS{Value: artistPK},
					":skPrefix": &types.AttributeValueMemberS{Value: "NAME"},
				},
				Limit: aws.Int32(1),
			})
			if aerr != nil || len(artistResp.Items) == 0 {
				continue
			}
			var artistRec struct {
				ArtistName string `dynamodbav:"artist_name"`
				ArtistBio  string `dynamodbav:"artist_bio"`
			}
			if err := attributevalue.UnmarshalMap(artistResp.Items[0], &artistRec); err != nil {
				continue
			}
			artistNames = append(artistNames, artistRec.ArtistName)
			artistBios = append(artistBios, artistRec.ArtistBio)
		}

		dto := &models.EventDTO{
			EventID:     eventID,
			EventName:   dbRec.EventName,
			Description: dbRec.Description,
			Duration:    dbRec.Duration,
			Category:    dbRec.Category,
			IsBlocked:   dbRec.IsBlocked,
			ArtistNames: artistNames,
			ArtistBios:  artistBios,
		}
		results = append(results, dto)
	}

	return results, nil
}

func (er *EventRepositoryDDB) GetEventsHostedByHost(ctx context.Context, hostID string) ([]models.EventDTO, error) {
	hostSK := "HOST#" + hostID

	var venueIDs []string

	scanInput := &dynamodb.ScanInput{
		TableName:        aws.String(er.TableName),
		FilterExpression: aws.String("begins_with(pk, :venuePrefix) AND sk = :hostSk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":venuePrefix": &types.AttributeValueMemberS{Value: "VENUE#"},
			":hostSk":      &types.AttributeValueMemberS{Value: hostSK},
		},
		ProjectionExpression: aws.String("pk"),
	}

	pg := dynamodb.NewScanPaginator(er.db, scanInput)
	for pg.HasMorePages() {
		out, err := pg.NextPage(ctx)
		if err != nil {
			log.Printf("scan for host venues failed: %v\n", err)
			return nil, err
		}
		for _, item := range out.Items {
			var rec struct {
				PK string `dynamodbav:"pk"`
			}
			if err := attributevalue.UnmarshalMap(item, &rec); err != nil {
				continue
			}
			if strings.HasPrefix(rec.PK, "VENUE#") {
				venueID := strings.TrimPrefix(rec.PK, "VENUE#")
				venueIDs = append(venueIDs, venueID)
			}
		}
	}

	if len(venueIDs) == 0 {
		return []models.EventDTO{}, nil
	}

	datePrefix := "DATE#"

	var venueClauses []string
	exprVals := map[string]types.AttributeValue{
		":datePrefix": &types.AttributeValueMemberS{Value: datePrefix},
	}
	for i, vid := range venueIDs {
		placeholder := fmt.Sprintf(":v%d", i)
		venueClauses = append(venueClauses, "venue_id = "+placeholder)
		exprVals[placeholder] = &types.AttributeValueMemberS{Value: vid}
	}
	venueFilter := strings.Join(venueClauses, " OR ")

	filterExpr := "begins_with(sk, :datePrefix) AND (" + venueFilter + ")"

	showScanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(er.TableName),
		FilterExpression:          aws.String(filterExpr),
		ExpressionAttributeValues: exprVals,
		ProjectionExpression:      aws.String("pk, sk, event_id, venue_id, show_id, show_date_time"),
	}

	eventIDsSet := make(map[string]struct{})

	showPg := dynamodb.NewScanPaginator(er.db, showScanInput)
	for showPg.HasMorePages() {
		out, err := showPg.NextPage(ctx)
		if err != nil {
			log.Printf("scan for shows failed: %v\n", err)
			return nil, err
		}
		for _, item := range out.Items {
			var showRec struct {
				PK          string `dynamodbav:"pk"`
				SK          string `dynamodbav:"sk"`
				EventID     string `dynamodbav:"event_id"`
				VenueID     string `dynamodbav:"venue_id"`
				ShowID      string `dynamodbav:"show_id"`
				ShowDateStr string `dynamodbav:"show_date_time"`
			}
			if err := attributevalue.UnmarshalMap(item, &showRec); err != nil {
				continue
			}
			eid := showRec.EventID
			if eid == "" {
				parts := strings.Split(showRec.PK, "#")
				if len(parts) >= 2 {
					eid = parts[1]
				}
			}
			if eid != "" {
				eventIDsSet[eid] = struct{}{}
			}
		}
	}

	if len(eventIDsSet) == 0 {
		return []models.EventDTO{}, nil
	}

	var results []models.EventDTO
	for eid := range eventIDsSet {
		dto, err := er.GetByID(ctx, eid)
		if err != nil {
			log.Printf("GetByID failed for event %s: %v\n", eid, err)
			continue
		}
		results = append(results, *dto)
	}

	return results, nil
}

func (er *EventRepositoryDDB) GetEventsByName(ctx context.Context, name string) ([]*models.EventDTO, error) {
	ids, err := er.SearchByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error from SearchByName: %s", err.Error())
	}
	if len(ids) == 0 {
		return []*models.EventDTO{}, nil
	}

	var events []*models.EventDTO

	for _, id := range ids {
		dto, err := er.GetByID(ctx, id)
		if err != nil {
			continue
		}
		events = append(events, dto)
	}
	return events, nil

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
		sk := item["sk"].(*types.AttributeValueMemberS).Value

		parts := strings.SplitN(sk, "EVENT_ID#", 2)
		if len(parts) != 2 {
			continue
		}

		eventID := parts[1]
		eventIDs = append(eventIDs, eventID)
	}

	return eventIDs, nil
}
