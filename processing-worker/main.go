package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	consumer "github.com/haijianyang/go-sqs-consumer"
)

type Request struct {
	Destination string `json:"destination"`
	Payload     string `json:"payload"`
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"delivery_error"`
}

func Handle(message *sqs.Message) error {
	fmt.Println("body: ", *message.Body)
	recvCountStr := message.Attributes["ApproximateReceiveCount"]
	fmt.Printf("ApproximateReceiveCount %s\n", *recvCountStr)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	payload, err := json.Marshal(Request{
		Destination: "https://e8808211e8f3.ngrok.app",
		Payload:     *message.Body,
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

	recvCount, err := strconv.ParseFloat(*recvCountStr, 64)
	if err != nil {
		fmt.Printf("something went wrong with float parsing %s", err.Error())
		return err
	}

	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(60)
	newVisibilityTimeout := int64(math.Pow(2, recvCount)) + 30 + int64(randomInt)

	_, err = sqsClient.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
		ReceiptHandle:     message.ReceiptHandle,
		QueueUrl:          urlResult.QueueUrl,
		VisibilityTimeout: &newVisibilityTimeout,
	})

	if err != nil {
		fmt.Printf("something went wrong changing message visibility %s\n", err.Error())
		return err
	}
	return fmt.Errorf("failed to deliver event %v", lambdaResp)
}

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	queueName := aws.String("webhooks-queue")
	sqsClient := sqs.New(sess)
	urlResult, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: queueName,
	})
	if err != nil {
		fmt.Printf("something went wrong getting queue URL %s\n", err.Error())
		return
	}

	worker := consumer.New(&consumer.Config{
		QueueUrl: urlResult.QueueUrl,
	}, sqsClient)

	go worker.Start(Handle)

	//worker.Concurrent(Handler, 6)

	select {}
}
