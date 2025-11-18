package userrepository

import (
	"context"
	"errors"
	"eventro_aws/internals/models"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserDDB struct {
	UserID      string      `dynamodbav:"user_id"`
	Username    string      `dynamodbav:"username"`
	Email       string      `dynamodbav:"pk"`
	SK          string      `dynamodbav:"sk"`
	PhoneNumber string      `dynamodbav:"phone_number"`
	Password    string      `dynamodbav:"password"`
	Role        models.Role `dynamodbav:"role"`
	IsBlocked   bool        `dynamodbav:"is_blocked"`
	Venue_ids   []string    `dynamodbav:"venue_ids,omitempty"`
}

type UserRepositoryDDB struct {
	db        *dynamodb.Client
	TableName string
}

func NewUserRepoDDB(db *dynamodb.Client, tableName string) *UserRepositoryDDB {
	return &UserRepositoryDDB{db: db, TableName: tableName}
}

func (ur UserRepositoryDDB) Create(user *models.User) error {

	item, err := attributevalue.MarshalMap(user)
	newAttributeKey := "sk"
	newAttributeValue := &types.AttributeValueMemberS{Value: "DETAILS"}
	pk := &types.AttributeValueMemberS{Value: "USER#" + user.Email}

	item["venue_ids"] = &types.AttributeValueMemberL{
		Value: []types.AttributeValue{},
	}

	item["pk"] = pk
	item[newAttributeKey] = newAttributeValue
	fmt.Println(item)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = ur.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(ur.TableName), Item: item,
	})
	if err != nil {
		log.Printf("could not add item to table. err: %v", err)
	}
	return err
}

func (ur UserRepositoryDDB) GetByEmail(email string) (*models.User, error) {
	user := models.User{Email: email}
	pk, err := attributevalue.Marshal("USER#" + user.Email)
	if err != nil {
		return nil, err
	}
	sk, err := attributevalue.Marshal("DETAILS")
	if err != nil {
		return nil, err
	}
	keyMap := map[string]types.AttributeValue{"pk": pk, "sk": sk}
	ctx := context.Background()
	response, err := ur.db.GetItem(ctx, &dynamodb.GetItemInput{
		Key: keyMap, TableName: aws.String(ur.TableName),
	})
	if len(response.Item) == 0 {
		return nil, errors.New("no user found")
	}

	if err != nil {
		log.Printf("Couldn't get info. Here's why: %v\n", err)
		return nil, err
	} else {
		err = attributevalue.UnmarshalMap(response.Item, &user)
		if err != nil {
			log.Printf("Couldn't unmarshal response. Here's why: %v\n", err)
		}
	}

	return &user, nil
}

func (ur UserRepositoryDDB) GetByID(id string) (*models.User, error) {
	return &models.User{}, nil
}

func (ur UserRepositoryDDB) Update(user *models.User) error {
	return nil
}

func (ur UserRepositoryDDB) Delete(id string) error {
	return nil

}

func (ur UserRepositoryDDB) GetBlockedUsers() ([]models.User, error) {
	return []models.User{}, nil

}

func (ur UserRepositoryDDB) GetUsers() ([]models.User, error) {
	return []models.User{}, nil
}
