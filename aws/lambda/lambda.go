package lambda

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

type Lambda struct {
	Tags map[string]*string
}

func FindFunction(lambdac aws.LambdaAPI, functionName string) (*Lambda, error) {
	output, err := lambdac.GetFunction(&lambda.GetFunctionInput{
		FunctionName: to.Strp(functionName), // Lambda ARN
	})

	if err != nil {
		return nil, err
	}

	return &Lambda{
		Tags: output.Tags,
	}, nil
}
