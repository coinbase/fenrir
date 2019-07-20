package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

// AWS::Lambda::Permission

func ValidateAWSLambdaPermission(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSLambdaPermission,
) error {
	ref, err := decodeRef(res.FunctionName)
	if err != nil || ref == "" {
		return resourceError(res, resourceName, "AWS::Lambda::Permission FunctionName must be !Ref")
	}

	return nil
}
