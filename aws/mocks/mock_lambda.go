package mocks

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/coinbase/fenrir/aws"
)

type LambdaClient struct {
	aws.LambdaAPI
}

func (m *LambdaClient) GetFunction(in *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
	projectName := "project"
	configName := "otherconfig"

	if *in.FunctionName == "valid_lambda_arn" {
		projectName = "project"
		configName = "development"
	}

	return &lambda.GetFunctionOutput{
		Tags: map[string]*string{
			"ProjectName": &projectName,
			"ConfigName":  &configName,
		},
	}, nil
}
