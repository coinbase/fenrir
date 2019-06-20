package template

import (
	"fmt"
	"strings"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/coinbase/odin/aws"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/utils/to"
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

	bucket, _, uri, err := ValidateS3FilePropertyValues(res.Properties["Bucket"], res.Properties["Key"], res.Properties["Uri"])
	if err != nil {
		return resourceError(res, resourceName, err.Error())
	}

	if !strings.HasPrefix(uri, "s3://") {
		return resourceError(res, resourceName, "Uri must start with s3://")
	}

	// Validate Bucket tags
	tags, err := s3.GetBucketTags(s3c, to.Strp(bucket))
	if err != nil {
		return err
	}

	if err := hasCorrectTags(projectName, configName, tags); err != nil {
		return err
	}

	// URI must be a s3:// uri
	if _, ok := s3shas[uri]; !ok {
		return resourceError(res, resourceName, "Uri must be in S3FileSHAs")
	}

	return nil
}

func ValidateS3FilePropertyValues(bucketI, keyI, uriI interface{}) (string, string, string, error) {
	// Validate presence (although schema should also validate)
	if bucketI == nil || keyI == nil || uriI == nil {
		return "", "", "", fmt.Errorf("Bucket, Uri and Key are required properties")
	}

	var bucket string
	var key string
	var uri string

	// Properties
	switch bucketI.(type) {
	case string:
		bucket = bucketI.(string)
	}

	switch keyI.(type) {
	case string:
		key = keyI.(string)
	}

	switch uriI.(type) {
	case string:
		uri = uriI.(string)
	}

	if bucket == "" || key == "" || uri == "" {
		return "", "", "", fmt.Errorf("key, uri, or bucket not stirng or unset")
	}

	return bucket, key, uri, nil
}
