package template

import (
	"github.com/awslabs/goformation/v3/cloudformation"
	"github.com/awslabs/goformation/v3/cloudformation/sqs"
)

// AWS::SQS::Queue

func ValidateAWSSQSQueue(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *sqs.Queue,
) error {

	if res.DeletionPolicy() == "" {
		res.SetDeletionPolicy("Retain")
	}

	if res.QueueName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.QueueName = normalizeName("fenrir", projectName, configName, resourceName, 255)

	return nil
}
