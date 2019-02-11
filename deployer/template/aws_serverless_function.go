package template

import (
	"fmt"

	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/iam"
	"github.com/coinbase/fenrir/aws/sg"
	"github.com/coinbase/fenrir/aws/subnet"
	"github.com/coinbase/step/utils/to"
	"github.com/grahamjenson/goformation/cloudformation"
	"github.com/grahamjenson/goformation/cloudformation/resources"
)

func ValidateAWSServerlessFunction(
	projectName, configName, resourceName string,
	template *cloudformation.Template,
	fun *resources.AWSServerlessFunction,
	s3shas map[string]string,
	iamc aws.IAMAPI,
	ec2c aws.EC2API,
	s3c aws.S3API,
	kinc aws.KINAPI,
	ddbc aws.DDBAPI,
	sqsc aws.SQSAPI,
) error {

	if fun.FunctionName != "" {
		return resourceError(fun, resourceName, "Names are overwritten")
	}

	// Forces Lambda name for no conflicts
	fun.FunctionName = normalizeName("fenrir", projectName, configName, resourceName)

	if fun.Tags == nil {
		fun.Tags = map[string]string{}
	}

	fun.Tags["ProjectName"] = projectName
	fun.Tags["ConfigName"] = configName
	fun.Tags["ServiceName"] = resourceName

	// Role Must be Name and NOT intrinsic
	// We make sure it exists and has the correct tags
	role, err := iam.GetRole(iamc, &fun.Role)
	if err != nil {
		return resourceError(fun, resourceName, fmt.Sprintf("%v %v", fun.Role, err.Error()))
	}

	fun.Role = to.Strs(role.Arn)

	if err := ValidateResource("Role", projectName, configName, resourceName, role); err != nil {
		return resourceError(fun, resourceName, err.Error())
	}

	if fun.VpcConfig != nil {
		if err := ValidateVPCConfig(projectName, configName, resourceName, fun, ec2c); err != nil {
			return resourceError(fun, resourceName, err.Error())
		}
	}

	// Support and Validate These Events
	// S3 SNS Kinesis DynamoDB SQS Api Schedule CloudWatchEvent CloudWatchLogs IoTRule AlexaSkill
	for eventName, event := range fun.Events {
		switch event.Type {
		case "Api":
			if err := ValidateAPIEvent(template, event.Properties.ApiEvent); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("API Event %q %v", eventName, err.Error()))
			}
		case "S3":
			if err := ValidateS3Event(projectName, configName, event.Properties.S3Event, s3c); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("S3 Event %q %v", eventName, err.Error()))
			}
		case "Kinesis":
			if err := ValidateKinesisEvent(projectName, configName, event.Properties.KinesisEvent, kinc); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("Kinesis Event %q %v", eventName, err.Error()))
			}
		case "DynamoDB":
			if err := ValidateDynamoDBEvent(projectName, configName, event.Properties.DynamoDBEvent, ddbc); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("DynamoDB Event %q %v", eventName, err.Error()))
			}
		case "SQS":
			if err := ValidateSQSEvent(projectName, configName, event.Properties.SQSEvent, sqsc); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("SQS Event %q %v", eventName, err.Error()))
			}
		case "Schedule":
			if err := ValidateScheduleEvent(event.Properties.ScheduleEvent); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("Schedule Event %q %v", eventName, err.Error()))
			}
		case "CloudWatchEvent":
			if err := ValidateCloudWatchEventEvent(event.Properties.CloudWatchEventEvent); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("CloudWatch Event %q %v", eventName, err.Error()))
			}
		default:
			return resourceError(fun, resourceName, fmt.Sprintf("Event %q Unsupported Event type %q", eventName, event.Type))
		}
	}

	// CodeURI checking
	if fun.CodeUri != nil {
		if fun.CodeUri.S3Location != nil {
			return resourceError(fun, resourceName, "CodeUri.S3Location not supported")
		}

		if fun.CodeUri.String == nil {
			return resourceError(fun, resourceName, "CodeUri nil")
		}

		s3URI := *fun.CodeUri.String
		if _, ok := s3shas[s3URI]; !ok {
			return fmt.Errorf("CodeUri %v not included in the SHA256s map", s3URI)
		}
	}

	return nil
}

func ValidateVPCConfig(
	projectName, configName, resourceName string,
	fun *resources.AWSServerlessFunction,
	ec2c aws.EC2API,
) error {
	if len(fun.VpcConfig.SecurityGroupIds) < 1 {
		return fmt.Errorf("VpcConfig No security groups defined")
	}

	if len(fun.VpcConfig.SubnetIds) < 1 {
		return fmt.Errorf("VpcConfig No Subnets defined")
	}

	// Security Groups
	sgs, err := sg.Find(ec2c, strA(fun.VpcConfig.SecurityGroupIds))
	if err != nil {
		return fmt.Errorf("VpcConfig Find Security Group Error %v", err.Error())
	}

	// replace
	ids := []string{}
	for _, securityGroup := range sgs {
		ids = append(ids, *securityGroup.GroupID)
		if err := ValidateResource("SecurityGroup", projectName, configName, resourceName, securityGroup); err != nil {
			return fmt.Errorf("VpcConfig %v", err.Error())
		}
	}

	fun.VpcConfig.SecurityGroupIds = ids // replace

	// Subnets
	subnets, err := subnet.Find(ec2c, strA(fun.VpcConfig.SubnetIds))
	if err != nil {
		return fmt.Errorf("VpcConfig Find Subnet Error %v", err.Error())
	}

	ids = []string{}
	for _, sub := range subnets {
		ids = append(ids, *sub.SubnetID)
		if err := ValidateSubnet(sub); err != nil {
			return fmt.Errorf("VpcConfig Validate Subnet Error %v", err.Error())
		}
	}

	fun.VpcConfig.SubnetIds = ids // replace

	return nil
}
