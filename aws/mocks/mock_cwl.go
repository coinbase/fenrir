package mocks

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

type CWLClient struct {
	aws.CWLAPI
}

// ListTagsLogGroup returns
func (m *CWLClient) ListTagsLogGroup(in *cloudwatchlogs.ListTagsLogGroupInput) (*cloudwatchlogs.ListTagsLogGroupOutput, error) {
	return &cloudwatchlogs.ListTagsLogGroupOutput{
		Tags: map[string]*string{
			"ProjectName": to.Strp("project"),
			"ConfigName":  to.Strp("development"),
		},
	}, nil
}
