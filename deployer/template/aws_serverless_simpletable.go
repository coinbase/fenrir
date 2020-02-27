package template

import (
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/policies"
	"github.com/awslabs/goformation/v4/cloudformation/serverless"
)

// AWS::Serverless::SimpleTable

func ValidateAWSServerlessSimpleTable(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *serverless.SimpleTable,
) error {

	if res.AWSCloudFormationDeletionPolicy == "" {
		res.AWSCloudFormationDeletionPolicy = policies.DeletionPolicy("Retain")
	}

	if res.TableName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.TableName = normalizeName("fenrir", projectName, configName, resourceName, 255)

	return nil
}
