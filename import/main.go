package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func sendMessageToQueue(message string, queueURL *string, svc *sqs.SQS) error {
	fmt.Printf("Message: %s QueueURL: %s", message, *queueURL)
	sendMessageInput := &sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Message": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String("Hello there"),
			},
		},
		MessageBody: aws.String(message),
		QueueUrl:    queueURL,
	}

	_, err := svc.SendMessage(sendMessageInput)

	if err != nil {
		fmt.Println("Error")
		fmt.Println(err)
	}

	return err
}

func handler(ctx context.Context) error {
	fmt.Println("In handler")
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		return errors.New("'QUEUE_NAME' is not defined")
	}
	fmt.Println("Value of queueName is", queueName)

	tableName := os.Getenv("TRACKING_TABLE_NAME")
	if tableName == "" {
		return errors.New("'TRACKING_TABLE_NAME' is not defined")
	}
	fmt.Println("Value of tableName is", tableName)

	dynamodbSvc := dynamodb.New(session.New())

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	fmt.Println("Got session")
	svc := sqs.New(sess)
	fmt.Println("Got new SQS session")
	urlResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		fmt.Println("Error getting queue URL:", err)
	}
	fmt.Println("URL result", urlResult)
	queueURL := urlResult.QueueUrl
	fmt.Println("Queue URL", queueURL)

	for i := 0; i < 20; i++ {
		input := &dynamodb.PutItemInput{
			Item: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(fmt.Sprint(i)),
				},
				"messageProcessed": {
					BOOL: aws.Bool(false),
				},
			},
			TableName: aws.String(tableName),
		}
		result, err := dynamodbSvc.PutItem(input)
		if err != nil {
			fmt.Println("Error putting item in dynamodb", err)
		}
		fmt.Println("Put item successfully", result)

		sendMessageToQueue(fmt.Sprint(i), queueURL, svc)
	}

	return nil
}

func main() {
	fmt.Println("Running handler")
	lambda.Start(handler)
}
