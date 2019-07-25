package template

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/subnet"
	"github.com/coinbase/step/utils/to"
)

func ValidateTemplateResources(
	projectName, configName, region, accountId string,
	template *cloudformation.Template,
	s3shas map[string]string,
	iamc aws.IAMAPI,
	ec2c aws.EC2API,
	s3c aws.S3API,
	kinc aws.KINAPI,
	ddbc aws.DDBAPI,
	sqsc aws.SQSAPI,
	snsc aws.SNSAPI,
	kmsc aws.KMSAPI,
	lambdac aws.LambdaAPI,
) error {

	for name, a := range template.Resources {
		switch a.AWSCloudFormationType() {
		case "AWS::Serverless::Function":
			res, err := template.GetAWSServerlessFunctionWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSServerlessFunction(
				projectName, configName, region, accountId, name,
				template, res, s3shas,
				iamc, ec2c, s3c, kinc, ddbc, sqsc, snsc, kmsc); err != nil {
				return err
			}
		case "AWS::Serverless::Api":
			res, err := template.GetAWSServerlessApiWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSServerlessApi(projectName, configName, name, template, res, s3shas); err != nil {
				return err
			}

		case "AWS::Serverless::LayerVersion":
			res, err := template.GetAWSServerlessLayerVersionWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSServerlessLayerVersion(projectName, configName, name, template, res, s3shas); err != nil {
				return err
			}

		case "AWS::Serverless::SimpleTable":
			res, err := template.GetAWSServerlessSimpleTableWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSServerlessSimpleTable(projectName, configName, name, template, res); err != nil {
				return err
			}

		case "AWS::SQS::Queue":
			res, err := template.GetAWSSQSQueueWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSSQSQueue(projectName, configName, name, template, res); err != nil {
				return err
			}

		case "AWS::CloudFront::Distribution":
			res, err := template.GetAWSCloudFrontDistributionWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSCloudFrontDistribution(projectName, configName, name, template, res); err != nil {
				return err
			}

		case "AWS::ElasticLoadBalancingV2::LoadBalancer":
			res, err := template.GetAWSElasticLoadBalancingV2LoadBalancerWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSElasticLoadBalancingV2LoadBalancer(projectName, configName, name, template, ec2c, res); err != nil {
				return err
			}

		case "AWS::ElasticLoadBalancingV2::TargetGroup":
			res, err := template.GetAWSElasticLoadBalancingV2TargetGroupWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSElasticLoadBalancingV2TargetGroup(projectName, configName, name, template, lambdac, res); err != nil {
				return err
			}

		case "AWS::ElasticLoadBalancingV2::Listener":
			res, err := template.GetAWSElasticLoadBalancingV2ListenerWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSElasticLoadBalancingV2Listener(projectName, configName, name, template, res); err != nil {
				return err
			}

		case "AWS::ElasticLoadBalancingV2::ListenerRule":
			res, err := template.GetAWSElasticLoadBalancingV2ListenerRuleWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSElasticLoadBalancingV2ListenerRule(projectName, configName, name, template, res); err != nil {
				return err
			}

		case "AWS::Lambda::Permission":
			res, err := template.GetAWSLambdaPermissionWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSLambdaPermission(projectName, configName, name, template, res); err != nil {
				return err
			}

		default:
			return fmt.Errorf("Unsupported type %q for %q", a.AWSCloudFormationType(), name)
		}
	}

	return nil
}

func ValidateSubnet(sub *subnet.Subnet) error {
	if sub.DeployWithFenrirTag == nil {
		return fmt.Errorf("DeployWithFenrir Tag is nil")
	}
	return nil
}

// UTILS
func ValidateResource(prefix, projectName, configName, serviceName string, res interface {
	ProjectName() *string
	ConfigName() *string
	ServiceName() *string
}) error {

	if !(aws.HasProjectName(res, &projectName) || aws.HasAllValue(res.ProjectName())) {
		return fmt.Errorf("Incorrect ProjectName for %v: has %q requires %q", prefix, to.Strs(res.ProjectName()), projectName)
	}

	if !(aws.HasConfigName(res, &configName) || aws.HasAllValue(res.ConfigName())) {
		return fmt.Errorf("Incorrect ConfigName for %v: has %q requires %q", prefix, to.Strs(res.ConfigName()), configName)
	}

	if !(aws.HasServiceName(res, &serviceName) || aws.HasAllValue(res.ServiceName())) {
		return fmt.Errorf("Incorrect ServiceName for %v: has %q requires %q", prefix, to.Strs(res.ServiceName()), serviceName)
	}

	return nil
}

func hasCorrectTags(projectName, configName string, tags map[string]string) error {
	if tags["ProjectName"] == projectName && tags["ConfigName"] == configName {
		return nil
	}

	if tags[fmt.Sprintf("FenrirAllowed:%v:%v", projectName, configName)] == "true" {
		return nil
	}

	if tags["FenrirAllAllowed"] == "true" {
		return nil
	}

	return fmt.Errorf("ProjectName (%v != %v) OR ConfigName (%v != %v) tags incorrect", tags["ProjectName"], projectName, tags["ConfigName"], configName)
}

func strA(strl []string) []*string {
	stra := []*string{}
	for _, s := range strl {
		stra = append(stra, to.Strp(s))
	}
	return stra
}

func normalizeName(prefix, projectName, configName, resourceName string, maxLength int) string {
	str := fmt.Sprintf("%v-%v-%v-%v", prefix, projectName, configName, resourceName)
	str = strings.Replace(str, "/", "-", -1)

	if len(str) > maxLength {
		digest := sha1.Sum([]byte(str))

		// Truncate to `maxLength` characters
		// Replace the last 8 characters with a digest (4 bytes = 8 hex chars)
		str = fmt.Sprintf("%s%x", str[:maxLength-8], digest[:4])
	}

	return str
}

func resourceError(resource cloudformation.Resource, name, errStr string) error {
	return fmt.Errorf("%v#%v: %v", resource.AWSCloudFormationType(), name, errStr)
}
