package showrepository

import (
	"context"

	"errors"
	"eventro_aws/internals/models"
	venuerepository "eventro_aws/internals/repository/venue_repository"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ShowRepositoryDDB struct {
	db        *dynamodb.Client
	TableName string
}

func NewShowRepositoryDDB(db *dynamodb.Client, tableName string) *ShowRepositoryDDB {
	return &ShowRepositoryDDB{db: db, TableName: tableName}
}

type ShowDDB struct {
	PK           string   `dynamodbav:"pk"`
	SK           string   `dynamodbav:"sk"`
	City         string   `dynamodbav:"venue_city"`
	VenueID      string   `dynamodbav:"venue_id"`
	EventID      string   `dynamodbav:"event_id"`
	CreatedAt    string   `dynamodbav:"created_at"`
	Price        int      `dynamodbav:"price"`
	ShowDateTime string   `dynamodbav:"show_date_time"`
	BookedSeats  []string `dynamodbav:"booked_seats"`
	IsBlocked    bool     `dynamodbav:"is_blocked"`
}

func (r *ShowRepositoryDDB) Create(ctx context.Context, show *models.Show) error {

	createdAt := show.CreatedAt.Format(time.RFC3339)
	showDateTime := show.ShowDate.Format("2006-01-02") + "T" + show.ShowTime
	venueRepo := venuerepository.NewVenueRepositoryDDB(r.db, r.TableName)
	venue, _ := venueRepo.GetByID(ctx, show.VenueID)
	city := venue.City

	showItem := map[string]any{
		"pk":             "SHOW#" + show.ID,
		"sk":             "DETAILS",
		"city":           city,
		"venue_id":       show.VenueID,
		"event_id":       show.EventID,
		"created_at":     createdAt,
		"price":          show.Price,
		"show_date_time": showDateTime,
		"booked_seats":   show.BookedSeats,
		"is_blocked":     show.IsBlocked,
	}

	// showItem := ShowDDB{
	// 	PK:             "SHOW#" + show.ID,
	// 	SK:             "DETAILS",
	// 	City:           city,
	// 	Venue_id:       show.VenueID,
	// 	event_id:       show.EventID,
	// 	created_at:     createdAt,
	// 	price:          show.Price,
	// 	show_date_time: showDateTime,
	// 	booked_seats:   show.BookedSeats,
	// 	is_blocked:     show.IsBlocked,
	// }

	av1, err := attributevalue.MarshalMap(showItem)
	if err != nil {
		return fmt.Errorf("failed to marshal show item: %w", err)
	}
	if _, err := r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName),
		Item:      av1,
	}); err != nil {
		return fmt.Errorf("failed to put SHOW item: %w", err)
	}

	// -----------------------------------------
	// 2) FETCH EVENT DETAILS FIRST
	// pk: EVENT#<event_id>, sk: DETAILS
	// -----------------------------------------
	eventPK := "EVENT#" + show.EventID

	evtOut, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: eventPK},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to fetch event details: %w", err)
	}
	if evtOut.Item == nil {
		return fmt.Errorf("event does not exist: %s", show.EventID)
	}

	var eventRec struct {
		EventName   string   `dynamodbav:"event_name"`
		Description string   `dynamodbav:"description"`
		Duration    string   `dynamodbav:"duration"`
		Category    string   `dynamodbav:"category"`
		IsBlocked   bool     `dynamodbav:"is_blocked"`
		ArtistIDs   []string `dynamodbav:"artist_ids"`
	}
	if err := attributevalue.UnmarshalMap(evtOut.Item, &eventRec); err != nil {
		return fmt.Errorf("failed to unmarshal event details: %w", err)
	}

	// -----------------------------------------
	// 3) CITY#<city> -> EVENT#<event_id>
	// -----------------------------------------
	cityEventItem := map[string]interface{}{
		"pk":          "CITY#" + city,
		"sk":          "EVENT#" + show.EventID,
		"event_name":  eventRec.EventName,
		"description": eventRec.Description,
		"duration":    eventRec.Duration,
		"category":    eventRec.Category,
		"is_blocked":  eventRec.IsBlocked,
		"artist_ids":  eventRec.ArtistIDs,
	}

	av2, err := attributevalue.MarshalMap(cityEventItem)
	if err != nil {
		return fmt.Errorf("failed to marshal city-event item: %w", err)
	}
	if _, err := r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName),
		Item:      av2,
	}); err != nil {
		return fmt.Errorf("failed to put CITY->EVENT item: %w", err)
	}

	// -----------------------------------------
	// 4) EVENT#<event_id>#CITY#<city> ->
	//    DATE#<date>#VENUE#<venue_id>#SHOW#<show_id>
	// -----------------------------------------
	pk3 := "EVENT#" + show.EventID + "#CITY#" + city
	sk3 := "DATE#" + showDateTime + "#VENUE#" + show.VenueID + "#SHOW#" + show.ID

	eventDateItem := map[string]interface{}{
		"pk":         pk3,
		"sk":         sk3,
		"is_blocked": show.IsBlocked,
		"price":      show.Price,
	}

	av3, err := attributevalue.MarshalMap(eventDateItem)
	if err != nil {
		return fmt.Errorf("failed to marshal event-date item: %w", err)
	}
	if _, err := r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName),
		Item:      av3,
	}); err != nil {
		return fmt.Errorf("failed to put EVENT#CITY->DATE item: %w", err)
	}

	return nil
}

func (r *ShowRepositoryDDB) GetByID(ctx context.Context, id string) (*models.ShowDTO, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	// 1. Fetch SHOW item
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "SHOW#" + id},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get show: %w", err)
	}
	if out.Item == nil {
		return nil, fmt.Errorf("show not found: %s", id)
	}

	// match DDB schema
	var showDDB struct {
		VenueID      string   `dynamodbav:"venue_id"`
		EventID      string   `dynamodbav:"event_id"`
		Price        float64  `dynamodbav:"price"`
		ShowDateTime string   `dynamodbav:"show_date_time"`
		BookedSeats  []string `dynamodbav:"booked_seats"`
	}
	if err := attributevalue.UnmarshalMap(out.Item, &showDDB); err != nil {
		return nil, fmt.Errorf("failed to unmarshal show: %w", err)
	}

	// Parse date/time
	parts := strings.Split(showDDB.ShowDateTime, "T")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid show_date_time: %s", showDDB.ShowDateTime)
	}
	date, _ := time.Parse("2006-01-02", parts[0])
	timeStr := parts[1]

	// 2. Fetch VENUE
	venueDTO, err := r.getVenueDTO(ctx, showDDB.VenueID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue: %w", err)
	}

	// 3. Construct ShowDTO
	return &models.ShowDTO{
		ID:          id,
		EventID:     showDDB.EventID,
		Price:       showDDB.Price,
		ShowDate:    date,
		ShowTime:    timeStr,
		BookedSeats: showDDB.BookedSeats,
		Venue:       *venueDTO,
	}, nil
}

func (r *ShowRepositoryDDB) ListByEvent(ctx context.Context, eventID, city, date, venueID string) ([]models.ShowDTO, error) {
	if eventID == "" || city == "" {
		return nil, errors.New("eventID and city are required")
	}

	pk := "EVENT#" + eventID + "#CITY#" + city

	skPrefix := "DATE#"
	if date != "" {
		skPrefix = "DATE#" + date
		if venueID != "" {
			skPrefix = skPrefix + "#VENUE#" + venueID
		}
	}

	// Query lightweight show rows
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
			":sk": &types.AttributeValueMemberS{Value: skPrefix},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query shows: %w", err)
	}

	shows := make([]models.ShowDTO, 0, len(out.Items))

	for _, item := range out.Items {

		// STEP A – Unmarshal lightweight row
		var row struct {
			PK        string  `dynamodbav:"pk"`
			SK        string  `dynamodbav:"sk"`
			Price     float64 `dynamodbav:"price"`
			IsBlocked bool    `dynamodbav:"is_blocked"`
		}
		if err := attributevalue.UnmarshalMap(item, &row); err != nil {
			return nil, fmt.Errorf("failed to unmarshal row: %w", err)
		}

		// Parse SK: DATE#<date>#VENUE#<venueID>#SHOW#<showID>
		parts := strings.Split(row.SK, "#Show#")
		// [DATE, <date>, VENUE, <venueID>, SHOW, <showID>]

		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid show SK format: %s", row.SK)
		}

		showID := parts[1]

		// STEP B – Fetch full SHOW details
		fullShow, err := r.GetByID(ctx, showID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch show details: %w", err)
		}

		// override price from row (if needed)
		fullShow.Price = row.Price

		shows = append(shows, *fullShow)
	}

	return shows, nil
}

func (r *ShowRepositoryDDB) getVenueDTO(ctx context.Context, VenueID string) (*models.VenueDTO, error) {
	venuePK := "VENUE#" + VenueID
	venueOut, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: venuePK},
			":sk": &types.AttributeValueMemberS{Value: "HOST#"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch venue: %w", err)
	}

	if len(venueOut.Items) == 0 {
		return nil, fmt.Errorf("venue not found: %s", VenueID)
	}
	var venueDDB models.VenueDTO
	if err := attributevalue.UnmarshalMap(venueOut.Items[0], &venueDDB); err != nil {
		return nil, fmt.Errorf("failed to unmarshal venue: %w", err)
	}
	return &venueDDB, nil
}

func (r *ShowRepositoryDDB) Update(ctx context.Context, showID string, isBlocked bool) error {
	if showID == "" {
		return errors.New("showID is required")
	}

	_, err := r.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "SHOW#" + showID},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
		UpdateExpression: aws.String("SET is_blocked = :b"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":b": &types.AttributeValueMemberBOOL{Value: isBlocked},
		},
		ReturnValues: types.ReturnValueNone,
	})

	if err != nil {
		return fmt.Errorf("failed to update show is_blocked: %w", err)
	}

	return nil
}

func (br *ShowRepositoryDDB) UpdateShowBooking(ctx context.Context, booking models.Booking) error {
	// STEP E: Update SHOW booked_seats
	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(br.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "SHOW#" + booking.ShowID},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
		UpdateExpression: aws.String(
			"SET booked_seats = list_append(if_not_exists(booked_seats, :empty), :newSeats)",
		),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":newSeats": &types.AttributeValueMemberL{
				Value: toAVList(booking.Seats),
			},
			":empty": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{},
			},
		},
		ReturnValues: types.ReturnValueNone,
	}

	_, err := br.db.UpdateItem(ctx, updateInput)
	if err != nil {
		return fmt.Errorf("failed to update show booked seats: %w", err)
	}
	return nil

}

func toAVList(strs []string) []types.AttributeValue {
	avs := make([]types.AttributeValue, 0, len(strs))
	for _, s := range strs {
		avs = append(avs, &types.AttributeValueMemberS{Value: s})
	}
	return avs
}
