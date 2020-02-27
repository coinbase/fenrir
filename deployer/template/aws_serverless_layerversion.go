package template

import (
	"fmt"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/serverless"
)

// AWS::Serverless::LayerVersion

func ValidateAWSServerlessLayerVersion(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *serverless.LayerVersion,
	s3shas map[string]string,
) error {

	if res.LayerName != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.LayerName = normalizeName("layer", projectName, configName, resourceName, 64)

	if res.ContentUri == "" {
		return resourceError(res, resourceName, "ContentUri is empty")
	}

	if _, ok := s3shas[res.ContentUri]; !ok {
		return fmt.Errorf("ContentUri %v not included in the SHA256s map", res.ContentUri)
	}

	return nil
}
