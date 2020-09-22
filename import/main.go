package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
		MessageBody: aws.String("General Kenobi"),
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

	for i := 0; i < 100; i++ {
		sendMessageToQueue("Hello there "+string(i), queueURL, svc)
	}

	return nil
}

func main() {
	fmt.Println("Running handler")
	lambda.Start(handler)
}
