package template

import (
	"fmt"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/iam"
	"github.com/coinbase/fenrir/aws/kms"
	"github.com/coinbase/fenrir/aws/sg"
	"github.com/coinbase/fenrir/aws/subnet"
	"github.com/coinbase/step/utils/to"
)

func ValidateAWSServerlessFunction(
	projectName, configName, region, accountId, resourceName string,
	template *cloudformation.Template,
	fun *resources.AWSServerlessFunction,
	s3shas map[string]string,
	iamc aws.IAMAPI,
	ec2c aws.EC2API,
	s3c aws.S3API,
	kinc aws.KINAPI,
	ddbc aws.DDBAPI,
	sqsc aws.SQSAPI,
	snsc aws.SNSAPI,
	kmsc aws.KMSAPI,
) error {

	if fun.FunctionName != "" {
		return resourceError(fun, resourceName, fmt.Sprintf("Names are overwritten, it is %v", fun.FunctionName))
	}

	// Forces Lambda name for no conflicts
	fun.FunctionName = normalizeName("fenrir", projectName, configName, resourceName, 64)

	if fun.Tags == nil {
		fun.Tags = map[string]string{}
	}

	fun.Tags["ProjectName"] = projectName
	fun.Tags["ConfigName"] = configName
	fun.Tags["ServiceName"] = resourceName

	if err := ValidateFunctionIAM(projectName, configName, accountId, resourceName, fun, iamc, kmsc); err != nil {
		return err
	}

	if fun.VpcConfig != nil {
		if err := ValidateVPCConfig(projectName, configName, resourceName, fun, ec2c); err != nil {
			return resourceError(fun, resourceName, err.Error())
		}
	}

	if err := ValidateFunctionEvents(template, projectName, configName, region, accountId, resourceName, fun, s3c, kinc, ddbc, sqsc, snsc); err != nil {
		return err
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

func ValidateFunctionIAM(
	projectName, configName, accountId, resourceName string,
	fun *resources.AWSServerlessFunction,
	iamc aws.IAMAPI,
	kmsc aws.KMSAPI,
) error {
	// IAM VALIDATIONS
	// Either Role XOR Policies

	if fun.Role != "" && fun.Policies == nil {

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

	} else if fun.Role == "" && fun.Policies != nil {
		fun.PermissionsBoundary = fmt.Sprintf("arn:aws:iam::%s:policy/fenrir-permissions-boundary", accountId)
		policies := fun.Policies
		if policies.String != nil ||
			policies.IAMPolicyDocument != nil {
			return resourceError(fun, resourceName, "Policies: only support SAMPolicyTemplateArray")
		}

		// Arrays are a bit annoying because they contain the zero values
		if policies.StringArray != nil {
			for _, s := range *policies.StringArray {
				if s != "" {
					return resourceError(fun, resourceName, fmt.Sprintf("Policies: only support SAMPolicyTemplateArray not StringArray with %q", s))
				}
			}
		}

		if policies.IAMPolicyDocumentArray != nil {
			for _, i := range *policies.IAMPolicyDocumentArray {
				if i.Statement != nil {
					return resourceError(fun, resourceName, "Policies: only support SAMPolicyTemplateArray not IAMPolicyDocumentArray")
				}
			}
		}

		if policies.SAMPolicyTemplateArray == nil || len(*policies.SAMPolicyTemplateArray) == 0 {
			return resourceError(fun, resourceName, "Policies: SAMPolicyTemplateArray undefined")
		}

		for _, p := range *policies.SAMPolicyTemplateArray {
			if p.DynamoDBCrudPolicy != nil {
				ref, err := decodeRef(p.DynamoDBCrudPolicy.TableName)
				if err != nil || ref == "" {
					return resourceError(fun, resourceName, "Policies.DynamoDBCrudPolicy.TableName must be !Ref")
				}
			} else if p.SQSPollerPolicy != nil {
				ref, err := decodeRef(p.SQSPollerPolicy.QueueName)
				if err != nil || ref == "" {
					return resourceError(fun, resourceName, "Policies.SQSPollerPolicy.QueueName must be !Ref")
				}
			} else if p.LambdaInvokePolicy != nil {
				ref, err := decodeRef(p.LambdaInvokePolicy.FunctionName)
				if err != nil || ref == "" {
					return resourceError(fun, resourceName, "Policies.LambdaInvokePolicy.FunctionName must be !Ref")
				}
			} else if p.KMSDecryptPolicy != nil {
				key, err := kms.FindKey(kmsc, p.KMSDecryptPolicy.KeyId)
				if err != nil {
					return resourceError(fun, resourceName, fmt.Sprintf("KMSDecryptPolicy %v", err.Error()))
				}

				// Overwrite keyID to be Key Id (in cases where it was set to an alias)
				p.KMSDecryptPolicy.KeyId = key.Id

				err = hasCorrectTags(projectName, configName, key.Tags)
				if err != nil {
					return resourceError(fun, resourceName, fmt.Sprintf("KMSDecryptPolicy %v", err.Error()))
				}
			} else if p.VPCAccessPolicy != nil {
				// All good
			} else {
				return resourceError(fun, resourceName, fmt.Sprintf("Policies: Unsupported SAMPolicyTemplate %s", to.CompactJSONStr(p)))
			}
		}

	} else {
		return resourceError(fun, resourceName, "Must define Role XOR Policies")
	}

	return nil
}

func ValidateFunctionEvents(
	template *cloudformation.Template,
	projectName, configName, region, accountId, resourceName string,
	fun *resources.AWSServerlessFunction,
	s3c aws.S3API,
	kinc aws.KINAPI,
	ddbc aws.DDBAPI,
	sqsc aws.SQSAPI,
	snsc aws.SNSAPI,
) error {
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
			if err := ValidateKinesisEvent(projectName, configName, region, accountId, event.Properties.KinesisEvent, kinc); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("Kinesis Event %q %v", eventName, err.Error()))
			}
		case "DynamoDB":
			if err := ValidateDynamoDBEvent(projectName, configName, event.Properties.DynamoDBEvent, ddbc); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("DynamoDB Event %q %v", eventName, err.Error()))
			}
		case "SQS":
			if err := ValidateSQSEvent(projectName, configName, region, accountId, event.Properties.SQSEvent, sqsc); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("SQS Event %q %v", eventName, err.Error()))
			}
		case "SNS":
			if err := ValidateSNSEvent(projectName, configName, region, accountId, event.Properties.SNSEvent, snsc); err != nil {
				return resourceError(fun, resourceName, fmt.Sprintf("SNS Event %q %v", eventName, err.Error()))
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
