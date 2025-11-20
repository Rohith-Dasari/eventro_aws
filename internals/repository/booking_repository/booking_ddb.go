package bookingrepository

import (
	"context"
	"eventro_aws/internals/models"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type BookingRepositoryDDB struct {
	db        *dynamodb.Client
	TableName string
}

type UserBookingDDB struct {
	UserEmail             string   `dynamodbav:"pk"`
	BookingDate_BookingID string   `dynamodbav:"sk"`
	ShowID                string   `dynamodbav:"show_id"`
	TimeBooked            string   `dynamodbav:"time_booked"`
	NumTicketsBooked      int      `dynamodbav:"num_tickets_booked"`
	TotalPrice            float64  `dynamodbav:"total_price"`
	Seats                 []string `dynamodbav:"seats"`
	VenueCity             string   `dynamodbav:"venue_city"`
	VenueName             string   `dynamodbav:"venue_name"`
	VenueState            string   `dynamodbav:"venue_state"`
	EventName             string   `dynamodbav:"event_name"`
	EventDuration         string   `dynamodbav:"event_duration"`
	EventID               string   `dynamodbav:"event_id"`
}

func NewBookingRepositoryDDB(db *dynamodb.Client, tableName string) *BookingRepositoryDDB {
	return &BookingRepositoryDDB{db: db, TableName: tableName}
}

func (br *BookingRepositoryDDB) Create(ctx context.Context, booking *models.Booking) error {
	showPK := "SHOW#" + booking.ShowID
	showSK := "DETAILS"

	showOut, err := br.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(br.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: showPK},
			"sk": &types.AttributeValueMemberS{Value: showSK},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to fetch show: %w", err)
	}
	if showOut.Item == nil {
		return fmt.Errorf("show not found: %s", booking.ShowID)
	}

	var showDDB struct {
		VenueID      string  `dynamodbav:"venue_id"`
		EventID      string  `dynamodbav:"event_id"`
		City         string  `dynamodbav:"city"`
		ShowDateTime string  `dynamodbav:"show_date_time"`
		Price        float64 `dynamodbav:"price"`
	}
	if err := attributevalue.UnmarshalMap(showOut.Item, &showDDB); err != nil {
		return err
	}

	venuePK := "VENUE#" + showDDB.VenueID

	venueOut, err := br.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(br.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: venuePK},
			":sk": &types.AttributeValueMemberS{Value: "HOST#"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("failed to query venue host: %w", err)
	}
	if len(venueOut.Items) == 0 {
		return fmt.Errorf("venue not found: %s", showDDB.VenueID)
	}

	var venueDDB struct {
		Name  string `dynamodbav:"venue_name"`
		City  string `dynamodbav:"venue_city"`
		State string `dynamodbav:"venue_state"`
	}
	if err := attributevalue.UnmarshalMap(venueOut.Items[0], &venueDDB); err != nil {
		return err
	}

	eventPK := "EVENT#" + showDDB.EventID

	eventOut, err := br.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(br.TableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: eventPK},
			"sk": &types.AttributeValueMemberS{Value: "DETAILS"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to fetch event: %w", err)
	}
	if eventOut.Item == nil {
		return fmt.Errorf("event not found: %s", showDDB.EventID)
	}

	var eventDDB struct {
		Name     string `dynamodbav:"event_name"`
		Duration string `dynamodbav:"duration"`
	}
	if err := attributevalue.UnmarshalMap(eventOut.Item, &eventDDB); err != nil {
		return err
	}

	compositeKey := "BOOKED_SHOW_DATE#" + showDDB.ShowDateTime + "#BOOKINGID#" + booking.BookingID

	pkValue := "USER#" + booking.UserID

	bookingDDB := UserBookingDDB{
		UserEmail:             pkValue,
		BookingDate_BookingID: compositeKey,
		ShowID:                booking.ShowID,
		TimeBooked:            booking.TimeBooked.String(),
		NumTicketsBooked:      booking.NumTickets,
		TotalPrice:            booking.TotalBookingPrice,
		Seats:                 booking.Seats,
		VenueCity:             venueDDB.City,
		VenueName:             venueDDB.Name,
		VenueState:            venueDDB.State,
		EventName:             eventDDB.Name,
		EventDuration:         eventDDB.Duration,
		EventID:               showDDB.EventID,
	}

	item, err := attributevalue.MarshalMap(bookingDDB)
	if err != nil {
		return err
	}

	_, err = br.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(br.TableName),
		Item:      item,
	})

	return err
}

func (r *BookingRepositoryDDB) ListByUser(ctx context.Context, userID string) ([]models.UserBookingDTO, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: "USER#" + userID},
			":skPrefix": &types.AttributeValueMemberS{Value: "BOOKED_SHOW_DATE#"},
		},
	}

	result, err := r.db.Query(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return []models.UserBookingDTO{}, nil
	}

	var bookingRecords []UserBookingDDB
	err = attributevalue.UnmarshalListOfMaps(result.Items, &bookingRecords)
	if err != nil {
		return nil, err
	}

	var dtoList []models.UserBookingDTO

	for _, b := range bookingRecords {
		parts := strings.Split(b.BookingDate_BookingID, "#")

		dto := models.UserBookingDTO{
			UserEmail:        b.UserEmail,
			BookingDate:      parts[1],
			BookingID:        parts[3],
			ShowID:           b.ShowID,
			TimeBooked:       b.TimeBooked,
			NumTicketsBooked: b.NumTicketsBooked,
			TotalPrice:       b.TotalPrice,
			Seats:            b.Seats,
			VenueCity:        b.VenueCity,
			VenueName:        b.VenueName,
			VenueState:       b.VenueState,
			EventName:        b.EventName,
			EventDuration:    b.EventDuration,
			EventID:          b.EventID,
		}

		dtoList = append(dtoList, dto)
	}

	if dtoList == nil {
		return []models.UserBookingDTO{}, nil
	}
	return dtoList, nil
}
