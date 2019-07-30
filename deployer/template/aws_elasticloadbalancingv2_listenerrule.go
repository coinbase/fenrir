package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

func ValidateAWSElasticLoadBalancingV2ListenerRule(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSElasticLoadBalancingV2ListenerRule,
) error {
	ref, err := decodeRef(res.ListenerArn)
	if err != nil || ref == "" {
		return resourceError(res, resourceName, "ListenerRule.ListenerArn must be !Ref")
	}

	for _, action := range res.Actions {
		if action.Type == "forward" {
			ref, err := decodeRef(action.TargetGroupArn)
			if err != nil || ref == "" {
				return resourceError(res, resourceName, "ListenerRule.Actions.TargetGroupArn must be !Ref")
			}
		}
	}

	return nil
}
