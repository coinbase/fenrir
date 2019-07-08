package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

// AWS::SQS::Queue

func ValidateAWSCloudFrontDistribution(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSCloudFrontDistribution,
) error {
	// Todo: What do we want to validate?
	return nil
}
