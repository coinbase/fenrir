package template

import (
	"fmt"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

func ValidateAWSServerlessApi(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSServerlessApi,
	s3shas map[string]string,
) error {

	if res.Name != "" {
		return resourceError(res, resourceName, "Names are overwritten")
	}

	res.Name = normalizeName("fenrir", projectName, configName, resourceName)

	// Change the default to private because DEFAULT PRIVATE
	if res.EndpointConfiguration == "" {
		res.EndpointConfiguration = "PRIVATE"
	}

	if res.EndpointConfiguration != "REGIONAL" && res.EndpointConfiguration != "EDGE" && res.EndpointConfiguration != "PRIVATE" {
		return resourceError(res, resourceName, "EndpointConfiguration must equal either REGIONAL EDGE PRIVATE")
	}

	if res.DefinitionUri != nil {
		if res.DefinitionUri.S3Location != nil {
			return resourceError(res, resourceName, "DefinitionUri.S3Location not supported")
		}
		if res.DefinitionUri.String == nil {
			return resourceError(res, resourceName, "DefinitionUri nil")
		}

		s3URI := *res.DefinitionUri.String
		if _, ok := s3shas[s3URI]; !ok {
			return resourceError(res, resourceName, fmt.Sprintf("DefinitionUri %v not included in the SHA256s map", s3URI))
		}
	}

	return nil
}
