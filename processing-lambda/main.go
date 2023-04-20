package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Request struct {
	Destination string `json:"destination"`
	Payload     string `json:"payload"`
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"delivery_error"`
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		// Update event status to processing

		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		payload, err := json.Marshal(Request{
			Destination: "https://e082fc03993a.ngrok.app",
			Payload:     message.Body,
		})
		if err != nil {
			fmt.Println("Error marshalling payload request")
			os.Exit(0)
		}

		lambdaClient := lambda.New(sess)
		result, err := lambdaClient.Invoke(&lambda.InvokeInput{
			FunctionName:   aws.String("webhooks-egress-lambda"),
			InvocationType: aws.String("RequestResponse"),
			Payload:        payload,
		})
		if err != nil {
			fmt.Printf("something went wrong invoking lambda %s\n", err.Error())
		}

		var lambdaResp Response
		if err = json.Unmarshal(result.Payload, &lambdaResp); err != nil {
			fmt.Printf("something went wrong unmarshling response %s\n", err.Error())
			return err
		}

		// After invoking
		// If it is successful
		// - Update success with DELIVERED and response
		// - Return nil

		if err == nil && lambdaResp.StatusCode >= 200 && lambdaResp.StatusCode < 300 {
			return nil
		}

		fmt.Printf("%#v\n", lambdaResp)

		// If it fails
		// - Update database with FAILED status and response
		// - Increase visibility timeout

		// Increasing visibility timeout
		queueName := aws.String("webhooks-queue")
		sqsClient := sqs.New(sess)
		urlResult, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: queueName,
		})
		if err != nil {
			fmt.Printf("something went wrong getting queue URL %s\n", err.Error())
			return err
		}
		queueURL := urlResult.QueueUrl

		recvCountStr := message.Attributes["ApproximateReceiveCount"]
		fmt.Printf("ApproximateReceiveCount %s\n", recvCountStr)
		recvCount, err := strconv.ParseFloat(recvCountStr, 64)
		if err != nil {
			fmt.Printf("something went wrong with float parsing %s", err.Error())
			return err
		}

		// https://cloud.google.com/storage/docs/retry-strategy
		rand.Seed(time.Now().UnixNano())
		randomInt := rand.Intn(60)
		newVisibilityTimeout := int64(math.Pow(2, recvCount)) + 30 + int64(randomInt)

		_, err = sqsClient.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
			ReceiptHandle:     &message.ReceiptHandle,
			QueueUrl:          queueURL,
			VisibilityTimeout: &newVisibilityTimeout,
		})

		if err != nil {
			fmt.Printf("something went wrong changing message visibility %s\n", err.Error())
			return err
		}
		return fmt.Errorf("failed to deliver event %v", lambdaResp)
	}

	return nil
}

// UnmarshalLambdaResponse extracts the json encoded error from `errorMessage` in the lambda response payload then unmarshal it to Error
func UnmarshalLambdaResponse(data []byte, jsonError Response) error {
	var lambdaErrPayload struct {
		ErrorMsg string `json:"errorMessage"`
	}

	if err := json.Unmarshal(data, &lambdaErrPayload); err != nil {
		return fmt.Errorf("unexpected lambda payload got: %s, err: %w", string(data), err)
	}

	if err := json.Unmarshal([]byte(lambdaErrPayload.ErrorMsg), jsonError); err != nil {
		return fmt.Errorf("unexpected lambda payload, got: %s, err: %w", string(data), err)
	}

	return nil
}

func main() {
	runtime.Start(handler)
}
