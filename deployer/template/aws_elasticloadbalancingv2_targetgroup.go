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

	// Only allow lambda targets for now
	if res.TargetType != "lambda" {
		return resourceError(res, resourceName, "TargetGroup.TargetType must be lambda")
	}

	// Only allow lambdas created in this template.
	// In the future we'll want to allow targets to be any lambda with correct tags
	// (tags specifically allowing this project or all fenrir projects)
	for _, target := range res.Targets {
		args, err := decodeGetAtt(target.Id)
		if err != nil || len(args) != 2 || args[1] != "Arn" {
			return resourceError(res, resourceName, "TargetGroup.Targets.Id must be \"!GetAtt <lambdaName> Arn\"")
		}
	}

	return nil
}
