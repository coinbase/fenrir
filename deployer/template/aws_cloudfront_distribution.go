package template

import (
	"github.com/awslabs/goformation/v3/cloudformation"
	"github.com/awslabs/goformation/v3/cloudformation/cloudfront"
	"github.com/awslabs/goformation/v3/cloudformation/tags"
)

// AWS::CloudFront::Distribution

func ValidateAWSCloudFrontDistribution(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *cloudfront.Distribution,
) error {
	res.Tags = append(res.Tags, tags.Tag{Key: "ProjectName", Value: projectName})
	res.Tags = append(res.Tags, tags.Tag{Key: "ConfigName", Value: configName})
	res.Tags = append(res.Tags, tags.Tag{Key: "ServiceName", Value: resourceName})

	// Disallow s3 origins for now - we need to validate them securely which isn't trivial
	for _, origin := range res.DistributionConfig.Origins {
		if origin.S3OriginConfig != nil {
			return resourceError(res, resourceName, "S3 Origins are not yet supported for cloudfront distributions")
		}
	}

	return nil
}
