package template

import (
	"fmt"
	"strings"
	"crypto/sha1"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/subnet"
	"github.com/coinbase/step/utils/to"
)

func ValidateTemplateResources(
	projectName, configName string,
	template *cloudformation.Template,
	s3shas map[string]string,
	iamc aws.IAMAPI,
	ec2c aws.EC2API,
	s3c aws.S3API,
	kinc aws.KINAPI,
	ddbc aws.DDBAPI,
	sqsc aws.SQSAPI,
	snsc aws.SNSAPI,
) error {

	for name, a := range template.Resources {
		switch a.AWSCloudFormationType() {
		case "AWS::Serverless::Function":
			res, err := template.GetAWSServerlessFunctionWithName(name)
			if err != nil {
				return err
			}

			if err := ValidateAWSServerlessFunction(projectName, configName, name, template, res, s3shas,
				iamc, ec2c, s3c, kinc, ddbc, sqsc, snsc); err != nil {
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

	if (len(str) > maxLength) {
		digest := sha1.Sum([]byte(str))

		// Truncate to `maxLength` characters
		// Replace the last 8 characters with a digest (4 bytes = 8 hex chars)
		str = fmt.Sprintf("%s%x", str[:maxLength - 8], digest[:4])
	}

	return str
}

func resourceError(resource cloudformation.Resource, name, errStr string) error {
	return fmt.Errorf("%v#%v: %v", resource.AWSCloudFormationType(), name, errStr)
}
