package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(func(_ interface{}) (interface{}, error) {
		return map[string]interface{}{
			"isBase64Encoded": false,
		    "statusCode": 200,
		    "statusDescription": "200 OK",
		    "headers": map[string]interface{}{
		        "Content-Type": "text/plain",
		    },
			"body": "Hello",
		}, nil
	})
}
