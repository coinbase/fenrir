package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

// AWS::CloudFront::Distribution

func ValidateAWSCloudFrontDistribution(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSCloudFrontDistribution,
) error {
	res.Tags = append(res.Tags, resources.Tag{Key: "ProjectName", Value: projectName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ConfigName", Value: configName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ServiceName", Value: resourceName})

	// Disallow s3 origins for now - we need to validate them securely which isn't trivial
	for _, origin := range res.DistributionConfig.Origins {
		if origin.S3OriginConfig != nil {
			return resourceError(res, resourceName, "S3 Origins are not yet supported for cloudfront distributions")
		}
	}

	return nil
}
