package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

// AWS::DynamoDB::Table

func ValidateAWSDynamoDBTable(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSDynamoDBTable,
) error {

	if res.TableName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.TableName = normalizeName("fenrir", projectName, configName, resourceName, 255)

	//if res.Tags == nil {
	//	res.Tags = make([]resources.Tag, 0)
	//}

	//res.Tags = append(res.Tags, resources.Tag{Key: "ProjectName", Value: projectName})
	//res.Tags = append(res.Tags, resources.Tag{Key: "ConfigName", Value: configName})
	//res.Tags = append(res.Tags, resources.Tag{Key: "ServiceName", Value: resourceName})

	// would potentially be nice to enable by default
	//    PointInTimeRecoverySpecification:
	//PointInTimeRecoveryEnabled: true

	return nil
}
