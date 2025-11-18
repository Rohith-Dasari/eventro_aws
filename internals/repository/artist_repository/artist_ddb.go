package artistrepository

import (
	"context"
	"eventro_aws/internals/models"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ArtistRepositoryDDB struct {
	db        *dynamodb.Client
	TableName string
}

func NewArtistRepositoryDDB(db *dynamodb.Client, tableName string) *ArtistRepositoryDDB {
	return &ArtistRepositoryDDB{db: db, TableName: tableName}
}

func (r *ArtistRepositoryDDB) Create(artist models.ArtistDTO) error {

	artist.ArtistID = "ARTIST#" + artist.ArtistID
	item, err := attributevalue.MarshalMap(artist)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.TableName), Item: item,
	})

	if err != nil {
		log.Printf("could not add item to table. err: %v", err)
	}
	return err
}

func (r *ArtistRepositoryDDB) GetByID(id string) (*models.ArtistDTO, error) {
	pk := "ARTIST#" + id
	ctx := context.Background()

	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: pk},
			":prefix": &types.AttributeValueMemberS{Value: "NAME"},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("artist query error: %w", err)
	}

	if len(out.Items) == 0 {
		return nil, fmt.Errorf("artist not found: %s", id)
	}

	var artist models.ArtistDTO
	if err := attributevalue.UnmarshalMap(out.Items[0], &artist); err != nil {
		return nil, fmt.Errorf("unmarshal artist error: %w", err)
	}

	return &artist, nil
}
