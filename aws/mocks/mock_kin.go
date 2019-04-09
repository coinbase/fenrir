package mocks

import (
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

type KINClient struct {
	aws.KINAPI
}

func (m *KINClient) ListTagsForStream(in *kinesis.ListTagsForStreamInput) (*kinesis.ListTagsForStreamOutput, error) {
	return &kinesis.ListTagsForStreamOutput{
		Tags: []*kinesis.Tag{
			&kinesis.Tag{Key: to.Strp("ProjectName"), Value: to.Strp("project")},
			&kinesis.Tag{Key: to.Strp("ConfigName"), Value: to.Strp("development")},
		},
	}, nil
}
