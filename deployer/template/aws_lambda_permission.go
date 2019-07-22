package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

func ValidateAWSLambdaPermission(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSLambdaPermission,
) error {
	if res.Action != "lambda:InvokeFunction" {
		return resourceError(res, resourceName, "Lambda::Permission.Action must be lambda:InvokeFunction")
	}

	// Currently we only allow permissions to grant access to ELB
	// We can add other things here as we need them
	if res.Principal != "elasticloadbalancing.amazonaws.com" {
		return resourceError(res, resourceName, "Lambda::Permission.Principal must be elasticloadbalancing.amazonaws.com")
	}

	args, err := decodeGetAtt(res.FunctionName)
	if err != nil || len(args) != 2 || args[1] != "Arn" {
		return resourceError(res, resourceName, "Lambda::Permission.FunctionName must be \"!GetAtt <lambdaName> Arn\"")
	}

	return nil
}
