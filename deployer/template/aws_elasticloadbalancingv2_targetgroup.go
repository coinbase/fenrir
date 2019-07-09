package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

func ValidateAWSElasticLoadBalancingV2TargetGroup(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSElasticLoadBalancingV2TargetGroup,
) error {
	res.Name = normalizeName("fenrir", projectName, configName, resourceName, 32)

	res.Tags = append(res.Tags, resources.Tag{Key: "ProjectName", Value: projectName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ConfigName", Value: configName})
	res.Tags = append(res.Tags, resources.Tag{Key: "ServiceName", Value: resourceName})

	return nil
}
