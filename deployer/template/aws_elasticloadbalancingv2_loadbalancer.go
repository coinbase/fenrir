package template

import (
	"fmt"

	"github.com/awslabs/goformation/v3/cloudformation"
	"github.com/awslabs/goformation/v3/cloudformation/elasticloadbalancingv2"

	"github.com/awslabs/goformation/v3/cloudformation/tags"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/sg"
	"github.com/coinbase/fenrir/aws/subnet"
)

func ValidateAWSElasticLoadBalancingV2LoadBalancer(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	ec2c aws.EC2API,
	res *elasticloadbalancingv2.LoadBalancer,
) error {
	res.Name = normalizeName("fenrir", projectName, configName, resourceName, 32)
	res.Tags = append(res.Tags, tags.Tag{Key: "ProjectName", Value: projectName})
	res.Tags = append(res.Tags, tags.Tag{Key: "ConfigName", Value: configName})
	res.Tags = append(res.Tags, tags.Tag{Key: "ServiceName", Value: resourceName})

	if res.Type != "application" {
		return resourceError(res, resourceName, "Only application load balancers are supported")
	}

	// Ipv4 is the default, so empty is fine too
	if res.IpAddressType != "" && res.IpAddressType != "ipv4" {
		return resourceError(res, resourceName, "Only ipv4 load balancers are supported")
	}

	if res.SecurityGroups != nil {
		if err := ValidateLoadbalancerSecurityGroups(projectName, configName, resourceName, res, ec2c); err != nil {
			return resourceError(res, resourceName, err.Error())
		}
	}

	if err := ValidateLoadbalancerSubnets(projectName, configName, resourceName, res, ec2c); err != nil {
		return resourceError(res, resourceName, err.Error())
	}

	return nil
}

func ValidateLoadbalancerSecurityGroups(
	projectName, configName, resourceName string,
	res *elasticloadbalancingv2.LoadBalancer,
	ec2c aws.EC2API,
) error {
	if len(res.SecurityGroups) < 1 {
		return fmt.Errorf("LoadBalancer No security groups defined")
	}

	// Security Groups
	sgs, err := sg.Find(ec2c, strA(res.SecurityGroups))
	if err != nil {
		return fmt.Errorf("LoadBalancer Find Security Group Error %v", err.Error())
	}

	// replace
	ids := []string{}
	for _, securityGroup := range sgs {
		ids = append(ids, *securityGroup.GroupID)
		if err := ValidateResource("SecurityGroup", projectName, configName, resourceName, securityGroup); err != nil {
			return fmt.Errorf("LoadBalancer%v", err.Error())
		}
	}

	res.SecurityGroups = ids // replace

	return nil
}

func ValidateLoadbalancerSubnets(
	projectName, configName, resourceName string,
	res *elasticloadbalancingv2.LoadBalancer,
	ec2c aws.EC2API,
) error {
	if len(res.Subnets) < 1 {
		return fmt.Errorf("LoadBalancer No Subnets defined")
	}

	// Subnets
	subnets, err := subnet.Find(ec2c, strA(res.Subnets))
	if err != nil {
		return fmt.Errorf("LoadBalancerFind Subnet Error %v", err.Error())
	}

	ids := []string{}
	for _, sub := range subnets {
		ids = append(ids, *sub.SubnetID)
		if err := ValidateSubnet(sub); err != nil {
			return fmt.Errorf("LoadBalancerValidate Subnet Error %v", err.Error())
		}
	}

	res.Subnets = ids // replace

	return nil
}
