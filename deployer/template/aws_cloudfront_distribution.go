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
	res.Tags = append(res.Tags, resources.Tag{Key: "ProjectName", Value: projectName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ConfigName", Value: configName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ServiceName", Value: resourceName})

	return nil
}
