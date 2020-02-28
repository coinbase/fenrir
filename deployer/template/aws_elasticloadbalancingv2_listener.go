package template

import (
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/elasticloadbalancingv2"
)

func ValidateAWSElasticLoadBalancingV2Listener(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	res *elasticloadbalancingv2.Listener,
) error {
	ref, err := decodeRef(res.LoadBalancerArn)
	if err != nil || ref == "" {
		return resourceError(res, resourceName, "Listener.LoadBalancerArn must be !Ref")
	}

	return nil
}
