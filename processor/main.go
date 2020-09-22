package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for index, message := range sqsEvent.Records {
		fmt.Printf("Number %s\n", message.Body)
		fmt.Printf("%d: The message %s for event source %s = %s \n", index, message.MessageId, message.EventSource, message.Body)
		for attributeName, attribute := range message.MessageAttributes {
			fmt.Printf("Message attribute: %s Value %s\n", attributeName, *attribute.StringValue)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
