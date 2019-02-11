package mocks

import (
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/aws/mocks"
)

// MockClients struct
type MockClients struct {
	S3Client  *mocks.MockS3Client
	CFClient  *CFClient
	EC2Client *EC2Client
	IAMClient *IAMClient
	SFNClient *mocks.MockSFNClient
	SNSClient *SNSClient
	KINClient *KINClient
	DDBClient *DDBClient
	SQSClient *SQSClient
}

// MockAWS mock clients
func MockAWS() *MockClients {
	return &MockClients{
		S3Client:  &mocks.MockS3Client{},
		CFClient:  &CFClient{},
		EC2Client: &EC2Client{},
		IAMClient: &IAMClient{},
		SFNClient: &mocks.MockSFNClient{},
		SNSClient: &SNSClient{},
		KINClient: &KINClient{},
		DDBClient: &DDBClient{},
		SQSClient: &SQSClient{},
	}
}

// S3Client returns
func (a *MockClients) S3(*string, *string, *string) aws.S3API {
	return a.S3Client
}

func (a *MockClients) CF(*string, *string, *string) aws.CFAPI {
	return a.CFClient
}

// EC2Client returns
func (a *MockClients) EC2(*string, *string, *string) aws.EC2API {
	return a.EC2Client
}

// IAMClient returns
func (a *MockClients) IAM(*string, *string, *string) aws.IAMAPI {
	return a.IAMClient
}

// SFNClient returns
func (a *MockClients) SFN(*string, *string, *string) aws.SFNAPI {
	return a.SFNClient
}

// SNSClient returns
func (a *MockClients) SNS(*string, *string, *string) aws.SNSAPI {
	return a.SNSClient
}

func (a *MockClients) KIN(*string, *string, *string) aws.KINAPI {
	return a.KINClient
}

func (a *MockClients) DDB(*string, *string, *string) aws.DDBAPI {
	return a.DDBClient
}

func (a *MockClients) SQS(*string, *string, *string) aws.SQSAPI {
	return a.SQSClient
}
