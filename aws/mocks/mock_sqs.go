package mocks

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

type SQSClient struct {
	aws.SQSAPI
}

func (m *SQSClient) ListQueueTags(in *sqs.ListQueueTagsInput) (*sqs.ListQueueTagsOutput, error) {
	return &sqs.ListQueueTagsOutput{
		Tags: map[string]*string{
			"ProjectName": to.Strp("project"),
			"ConfigName":  to.Strp("development"),
		},
	}, nil
}
