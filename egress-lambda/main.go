package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Destination string `json:"destination"`
	Payload     string `json:"payload"`
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"delivery_error"`
}

func handler(ctx context.Context, request Request) (Response, error) {
	fmt.Printf("Function is called %#v", request)
	r, err := http.NewRequest("POST", request.Destination, bytes.NewBuffer([]byte(request.Payload)))
	if err != nil {
		fmt.Printf("something went wrong %s\n", err.Error())
		return Response{}, err
	}
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	defer res.Body.Close()
	if err != nil {
		fmt.Printf("something went wrong 2 %s\n", err.Error())
		return Response{}, err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		fmt.Printf("something went wrong 3 %s\n", err.Error())
		return Response{}, err
	}

	return Response{
		StatusCode: res.StatusCode,
		Body:       buf.String(),
	}, nil
}

func main() {
	lambda.Start(handler)
}
