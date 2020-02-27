package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	ar "github.com/coinbase/step/aws"
)

// FetchEc2Tag extracts tags
func FetchEc2Tag(tags []*ec2.Tag, tagKey *string) *string {
	if tagKey == nil {
		return nil
	}

	for _, tag := range tags {
		if tag.Key == nil {
			continue
		}
		if *tag.Key == *tagKey {
			return tag.Value
		}
	}

	return nil
}

// HasAllValue checks for the _all value tag
func HasAllValue(tag *string) bool {
	if tag == nil {
		return false
	}
	return "_all" == *tag
}

// HasProjectName checks value
func HasProjectName(r interface {
	ProjectName() *string
}, projectName *string) bool {
	if r.ProjectName() == nil || projectName == nil {
		return false
	}
	return *r.ProjectName() == *projectName
}

// HasConfigName checks value
func HasConfigName(r interface {
	ConfigName() *string
}, configName *string) bool {
	if r.ConfigName() == nil || configName == nil {
		return false
	}
	return *r.ConfigName() == *configName
}

// HasServiceName checks value
func HasServiceName(r interface {
	ServiceName() *string
}, serviceName *string) bool {
	if r.ServiceName() == nil || serviceName == nil {
		return false
	}
	return *r.ServiceName() == *serviceName
}

// S3API aws API
type S3API s3iface.S3API

// CFAPI is cloudfomration API
type CFAPI cloudformationiface.CloudFormationAPI

// CWLAPI is cloud watch logs API
type CWLAPI cloudwatchlogsiface.CloudWatchLogsAPI

// EC2API aws API
type EC2API ec2iface.EC2API

// IAMAPI aws API
type IAMAPI iamiface.IAMAPI

// SFNAPI aws API
type SFNAPI sfniface.SFNAPI

// SNSAPI aws API
type SNSAPI snsiface.SNSAPI

// KINAPI kinesis api
type KINAPI kinesisiface.KinesisAPI

// DDBAPI DynamoDB api
type DDBAPI dynamodbiface.DynamoDBAPI

// SQSAPI SQS api
type SQSAPI sqsiface.SQSAPI

// KMSAPI KMS api
type KMSAPI kmsiface.KMSAPI

// LambdaAPI aws api
type LambdaAPI lambdaiface.LambdaAPI

// Clients for AWS
type Clients interface {
	S3(region *string, accountID *string, role *string) S3API
	CF(region *string, accountID *string, role *string) CFAPI
	EC2(region *string, accountID *string, role *string) EC2API
	IAM(region *string, accountID *string, role *string) IAMAPI
	SFN(region *string, accountID *string, role *string) SFNAPI
	SNS(region *string, accountID *string, role *string) SNSAPI
	KIN(region *string, accountID *string, role *string) KINAPI
	DDB(region *string, accountID *string, role *string) DDBAPI
	SQS(region *string, accountID *string, role *string) SQSAPI
	KMS(region *string, accountID *string, role *string) KMSAPI
	Lambda(region *string, accountID *string, role *string) LambdaAPI
	CWL(region *string, accountID *string, role *string) CWLAPI
}

// ClientsStr implementation
type ClientsStr struct {
	ar.Clients
}

// S3 returns client for region account and role
func (awsc *ClientsStr) S3(region *string, accountID *string, role *string) S3API {
	return s3.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// CF returns cloudformation client
func (awsc *ClientsStr) CF(region *string, accountID *string, role *string) CFAPI {
	return cloudformation.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// CF returns cloudformation client
func (awsc *ClientsStr) CWL(region *string, accountID *string, role *string) CWLAPI {
	return cloudwatchlogs.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// EC2 returns client for region account and role
func (awsc *ClientsStr) EC2(region *string, accountID *string, role *string) EC2API {
	return ec2.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// IAM returns client for region account and role
func (awsc *ClientsStr) IAM(region *string, accountID *string, role *string) IAMAPI {
	return iam.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// SFN returns client for region account and role
func (awsc *ClientsStr) SFN(region *string, accountID *string, role *string) SFNAPI {
	return sfn.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// SNS returns client for region account and role
func (awsc *ClientsStr) SNS(region *string, accountID *string, role *string) SNSAPI {
	return sns.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// KIN returns client
func (awsc *ClientsStr) KIN(region *string, accountID *string, role *string) KINAPI {
	return kinesis.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// DDB returns client
func (awsc *ClientsStr) DDB(region *string, accountID *string, role *string) DDBAPI {
	return dynamodb.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// SQS returns client
func (awsc *ClientsStr) SQS(region *string, accountID *string, role *string) SQSAPI {
	return sqs.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// KMS returns client
func (awsc *ClientsStr) KMS(region *string, accountID *string, role *string) KMSAPI {
	return kms.New(awsc.Session(), awsc.Config(region, accountID, role))
}

// LAMBDA returns client
func (awsc *ClientsStr) Lambda(region *string, accountID *string, role *string) LambdaAPI {
	return lambda.New(awsc.Session(), awsc.Config(region, accountID, role))
}
