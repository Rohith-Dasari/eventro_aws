package venuerepository

import (
	"context"
	"errors"
	"eventro_aws/internals/models"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type VenueRepositoryDDB struct {
	db        *dynamodb.Client
	tableName string
}

func NewVenueRepositoryDDB(db *dynamodb.Client, tableName string) *VenueRepositoryDDB {

	return &VenueRepositoryDDB{db: db, tableName: tableName}
}

func (r *VenueRepositoryDDB) Create(ctx context.Context, venue *models.Venue) error {

	// 1) INSERT VENUE ITEM
	venueItem := map[string]interface{}{
		"pk":          "VENUE#" + venue.ID,
		"sk":          "HOST#" + venue.HostID,
		"venue_name":  venue.Name,
		"is_blocked":  venue.IsBlocked,
		"venue_city":  venue.City,
		"venue_state": venue.State,
	}

	itemAV, err := attributevalue.MarshalMap(venueItem)
	if err != nil {
		return fmt.Errorf("marshal venue: %w", err)
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      itemAV,
	})
	if err != nil {
		return fmt.Errorf("put venue failed: %w", err)
	}

	// 2) UPDATE USER VENUE LIST
	userPK := "USER#" + venue.HostID

	// Push venue ID into venue_ids list
	_, err = r.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: userPK},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
		UpdateExpression: aws.String("SET venue_ids = list_append(if_not_exists(venue_ids, :emptyList), :v)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":v":         &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberS{Value: venue.ID}}},
			":emptyList": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
	})
	if err != nil {
		return fmt.Errorf("update user venue_ids failed: %w", err)
	}

	return nil
}

func (r *VenueRepositoryDDB) GetByID(ctx context.Context, id string) (*models.Venue, error) {

	pk := "VENUE#" + id

	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: pk},
			":skPrefix": &types.AttributeValueMemberS{Value: "HOST#"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("query venue failed: %w", err)
	}

	if len(out.Items) == 0 {
		return nil, errors.New("venue not found")
	}

	// Unmarshal into Venue
	var venue models.Venue
	if err := attributevalue.UnmarshalMap(out.Items[0], &venue); err != nil {
		return nil, fmt.Errorf("unmarshal venue failed: %w", err)
	}

	venue.ID = id
	venue.HostID = strings.TrimPrefix(venue.HostID, "HOST#")

	return &venue, nil
}

func (r *VenueRepositoryDDB) ListByHost(ctx context.Context, hostID string) ([]models.Venue, error) {

	userIDs, err := r.getUserVenueIDs(hostID)
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		return []models.Venue{}, nil
	}

	keys := make([]map[string]types.AttributeValue, 0, len(userIDs))
	hostPK := "HOST#" + hostID

	for _, vid := range userIDs {
		keys = append(keys, map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "VENUE#" + vid},
			"sk": &types.AttributeValueMemberS{Value: hostPK},
		})
	}

	batchOut, err := r.db.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			r.tableName: {Keys: keys},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("batch get venues failed: %w", err)
	}

	items := batchOut.Responses[r.tableName]
	venues := make([]models.Venue, 0, len(items))

	for _, item := range items {

		var raw struct {
			PK        string `dynamodbav:"pk"`
			SK        string `dynamodbav:"sk"`
			Name      string `dynamodbav:"name"`
			City      string `dynamodbav:"city"`
			State     string `dynamodbav:"state"`
			IsBlocked bool   `dynamodbav:"is_blocked"`
		}

		if err := attributevalue.UnmarshalMap(item, &raw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal venue: %w", err)
		}

		venueID := strings.TrimPrefix(raw.PK, "VENUE#")
		hostID := strings.TrimPrefix(raw.SK, "HOST#")

		venues = append(venues, models.Venue{
			ID:        venueID,
			HostID:    hostID,
			Name:      raw.Name,
			City:      raw.City,
			State:     raw.State,
			IsBlocked: raw.IsBlocked,
		})
	}

	return venues, nil
}

func (r *VenueRepositoryDDB) getUserVenueIDs(email string) ([]string, error) {
	ctx := context.Background()

	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "USER#" + email},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
	})
	if err != nil {
		return nil, err
	}
	if out.Item == nil {
		return nil, fmt.Errorf("user not found")
	}

	var data struct {
		VenueIDs []string `dynamodbav:"venue_ids"`
	}
	if err := attributevalue.UnmarshalMap(out.Item, &data); err != nil {
		return nil, err
	}

	return data.VenueIDs, nil
}

func (r *VenueRepositoryDDB) Update(ctx context.Context, venue *models.Venue) error {
	item := map[string]interface{}{
		"pk":         "VENUE#" + venue.ID,
		"sk":         "HOST#" + venue.HostID,
		"name":       venue.Name,
		"city":       venue.City,
		"state":      venue.State,
		"is_blocked": venue.IsBlocked,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal venue: %w", err)
	}

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to update venue: %w", err)
	}

	return nil
}

func (r *VenueRepositoryDDB) Delete(ctx context.Context, id string) error {
	pk := "VENUE#" + id

	qOut, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
		},
	})
	if err != nil {
		return fmt.Errorf("query venue failed: %w", err)
	}
	if len(qOut.Items) == 0 {
		return fmt.Errorf("venue not found")
	}

	var record struct {
		SK string `dynamodbav:"sk"`
	}
	if err := attributevalue.UnmarshalMap(qOut.Items[0], &record); err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}
	hostEmail := strings.TrimPrefix(record.SK, "HOST#")

	_, err = r.db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: pk},
			"sk": &types.AttributeValueMemberS{Value: record.SK},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete venue: %w", err)
	}

	if err := r.removeVenueFromUser(hostEmail, id); err != nil {
		return fmt.Errorf("failed to remove venue from user: %w", err)
	}

	return nil
}
func (r *VenueRepositoryDDB) removeVenueFromUser(hostEmail, venueID string) error {
	ctx := context.Background()

	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "USER#" + hostEmail},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to read user: %w", err)
	}
	if out.Item == nil {
		return fmt.Errorf("user not found")
	}

	var user struct {
		VenueIDs []string `dynamodbav:"venue_ids"`
	}
	if err := attributevalue.UnmarshalMap(out.Item, &user); err != nil {
		return fmt.Errorf("failed to unmarshal user venue_ids: %w", err)
	}

	newVenueIDs := make([]string, 0, len(user.VenueIDs))
	for _, id := range user.VenueIDs {
		if id != venueID {
			newVenueIDs = append(newVenueIDs, id)
		}
	}

	marshalled, err := attributevalue.Marshal(newVenueIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal updated venue_ids: %w", err)
	}
	out.Item["venue_ids"] = marshalled

	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      out.Item,
	})
	if err != nil {
		return fmt.Errorf("failed to update user venue_ids: %w", err)
	}

	return nil
}
