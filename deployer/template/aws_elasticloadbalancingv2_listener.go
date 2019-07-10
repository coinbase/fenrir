package template

import (
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
)

func ValidateAWSElasticLoadBalancingV2Listener(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *resources.AWSElasticLoadBalancingV2Listener,
) error {
	ref, err := decodeRef(res.LoadBalancerArn)
	if err != nil || ref == "" {
		return resourceError(res, resourceName, "LoadbalancerListener.LoadBalancerArn must be !Ref")
	}

	return nil
}
