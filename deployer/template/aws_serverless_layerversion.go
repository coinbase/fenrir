package template

import (
	"fmt"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

// AWS::Serverless::LayerVersion

func ValidateAWSServerlessLayerVersion(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSServerlessLayerVersion,
	s3shas map[string]string,
) error {

	if res.LayerName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.LayerName = normalizeName("layer", projectName, configName, resourceName)

	if res.ContentUri == "" {
		return resourceError(res, resourceName, "ContentUri is empty")
	}

	if _, ok := s3shas[res.ContentUri]; !ok {
		return fmt.Errorf("ContentUri %v not included in the SHA256s map", res.ContentUri)
	}

	return nil
}
