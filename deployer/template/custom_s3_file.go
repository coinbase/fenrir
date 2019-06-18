package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/coinbase/odin/aws"
)

// AWS::Serverless::LayerVersion

func ValidateCustomS3File(
	projectName, configName, resourceName, lambdaArn string,
	template *cloudformation.Template,
	res *cloudformation.CustomResource,
	s3shas map[string]string,
	s3c aws.S3API,
) error {

	// We override the ServiceToken to be the Fenrir Lambda ARN
	if res.Properties["ServiceToken"] != nil {
		return resourceError(res, resourceName, "ServiceToken are overwritten")
	}

	res.Properties["ServiceToken"] = lambdaArn

	// TODO: Validate the S3Shas hash

	// TODO: Validate the Bucket has correct Tags

	// if res.ContentUri == "" {
	// 	return resourceError(res, resourceName, "ContentUri is empty")
	// }

	// if _, ok := s3shas[res.ContentUri]; !ok {
	// 	return fmt.Errorf("ContentUri %v not included in the SHA256s map", res.ContentUri)
	// }

	return nil
}
