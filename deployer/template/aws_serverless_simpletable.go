package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

// AWS::Serverless::SimpleTable

func ValidateAWSServerlessSimpleTable(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSServerlessSimpleTable,
) error {

	if res.TableName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.TableName = normalizeName("fenrir", projectName, configName, resourceName)

	return nil
}
