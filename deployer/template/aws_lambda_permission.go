package template

import (
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/lambda"
)

func ValidateAWSLambdaPermission(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *lambda.Permission,
) error {
	if res.Action != "lambda:InvokeFunction" {
		return resourceError(res, resourceName, "Lambda::Permission.Action must be lambda:InvokeFunction")
	}

	allowedPrincipals := []string{
		"elasticloadbalancing.amazonaws.com",
		"secretsmanager.amazonaws.com",
		"apigateway.amazonaws.com",
		"events.amazonaws.com",
	}

	if !(inSlice(res.Principal, allowedPrincipals)) {
		return resourceError(res, resourceName, res.Principal+" is not a currently supported value for Lambda::Permission.Principal")
	}

	args, err := decodeGetAtt(res.FunctionName)
	if err != nil || len(args) != 2 || args[1] != "Arn" {
		return resourceError(res, resourceName, "Lambda::Permission.FunctionName must be \"!GetAtt <lambdaName> Arn\"")
	}

	return nil
}

func inSlice(str string, slice []string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
