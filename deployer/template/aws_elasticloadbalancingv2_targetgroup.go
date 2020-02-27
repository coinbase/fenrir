package template

import (
	"fmt"

	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/elasticloadbalancingv2"
	"github.com/awslabs/goformation/v4/cloudformation/tags"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/lambda"
)

func ValidateAWSElasticLoadBalancingV2TargetGroup(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	lambdac aws.LambdaAPI,
	res *elasticloadbalancingv2.TargetGroup,
) error {
	res.Name = normalizeName("fenrir", projectName, configName, resourceName, 32)

	res.Tags = append(res.Tags, tags.Tag{Key: "ProjectName", Value: projectName})
	res.Tags = append(res.Tags, tags.Tag{Key: "ConfigName", Value: configName})
	res.Tags = append(res.Tags, tags.Tag{Key: "ServiceName", Value: resourceName})

	if res.TargetType == "instance" {
		// Currently only allow empty targets list - this allows ASGs to attach targets
		// but no individual instances can be added.
		if len(res.Targets) != 0 {
			return resourceError(res, resourceName, "TargetGroup.Targets must be empty for TargetType instance")
		}

		return nil
	}

	// Only allow lambda targets for now
	if res.TargetType != "lambda" {
		return resourceError(res, resourceName, "TargetGroup.TargetType must be lambda")
	}

	for _, target := range res.Targets {
		// Target can either be a lambda defined in this template (using !GetAtt to get the arn)
		// Or it can be a function name, arn, or partial arn of a lambda with the correct fenrir tags.
		args, err := decodeGetAtt(target.Id)
		if err != nil || len(args) != 2 || args[1] != "Arn" {
			lambda, err := lambda.FindFunction(lambdac, target.Id)
			if err != nil {
				return resourceError(res, resourceName, "TargetGroup.Targets.Id must be \"!GetAtt <lambdaName>.Arn\" or a valid lambda ARN")
			}

			if err := hasCorrectTags(projectName, configName, convTagMap(lambda.Tags)); err != nil {
				return resourceError(res, resourceName, fmt.Sprintf("TargetGroup.Target %v", err.Error()))
			}
		}
	}

	return nil
}

func convTagMap(tags map[string]*string) map[string]string {
	newTags := map[string]string{}

	for k, v := range tags {
		if v == nil {
			newTags[k] = ""
		} else {
			newTags[k] = *v
		}
	}

	return newTags
}
