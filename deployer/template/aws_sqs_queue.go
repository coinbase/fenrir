package template

import (
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/policies"
	"github.com/awslabs/goformation/v4/cloudformation/sqs"
)

// AWS::SQS::Queue

func ValidateAWSSQSQueue(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *sqs.Queue,
) error {

	if res.AWSCloudFormationDeletionPolicy == "" {
		res.AWSCloudFormationDeletionPolicy = policies.DeletionPolicy("Retain")
	}

	if res.QueueName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.QueueName = normalizeName("fenrir", projectName, configName, resourceName, 255)

	return nil
}
