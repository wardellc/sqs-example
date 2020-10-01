package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Item struct {
	id               string
	messageProcessed bool
}

func alreadyProcessed(svc *dynamodb.DynamoDB, tableName string, id string) (bool, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(tableName),
	}
	result, err := svc.GetItem(input)

	if err != nil {
		return true, errors.New("Error getting record " + id)
	}

	if result.Item == nil {
		return true, errors.New("Could not find item with id " + id)
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	messageProcessed := item.messageProcessed
	return messageProcessed, nil
}

func setMessageAsProcessed(svc *dynamodb.DynamoDB, tableName string, id string) error {
	// Update field to set messageProcessed to true
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":messageProcessed": {
				BOOL: aws.Bool(true),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		UpdateExpression: aws.String("set messageProcessed = :messageProcessed"),
		ReturnValues:     aws.String("UPDATED_NEW"),
		TableName:        aws.String(tableName),
	}
	response, err := svc.UpdateItem(input)
	fmt.Println("Set message as processed", id, "response", response, "error", err)
	return err
}

func processMessage(dynamodbSvc *dynamodb.DynamoDB, tableName string, id string) error {
	messageProcessed, err := alreadyProcessed(dynamodbSvc, tableName, id)
	if err != nil {
		return err
	}
	if !messageProcessed {
		fmt.Println("ID not already processed", id)
		randomSeed := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNumber := randomSeed.Intn(10)

		// 20% chance of erroring
		if randomNumber <= 1 {
			return errors.New("Mimicing error processing id " + id)
		}

		err = setMessageAsProcessed(dynamodbSvc, tableName, id)
		if err != nil {
			fmt.Println("Processed successfully but failed to update messageProcessed field for id", id)
		}
		fmt.Println("ID successfully processed", id)
	} else {
		fmt.Println("ID processed already", id)
	}
	return nil
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	fmt.Println("SQS event message is:", sqsEvent.Records)
	fmt.Println("Batch is of size", len(sqsEvent.Records))

	tableName := os.Getenv("TRACKING_TABLE_NAME")
	if tableName == "" {
		return errors.New("'TRACKING_TABLE_NAME' is not defined")
	}

	dynamodbSvc := dynamodb.New(session.New())

	for _, message := range sqsEvent.Records {
		id := message.Body
		err := processMessage(dynamodbSvc, tableName, id)
		if err != nil {
			fmt.Println("Error processing id", id, "with error", err)
			return err
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
