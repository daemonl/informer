package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct{}

func Handler(request Request) (interface{}, error) {

	return nil, nil
}

func main() {
	lambda.Start(Handler)
}
