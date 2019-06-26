package mocks

import (
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

type KMSClient struct {
	aws.KMSAPI
}

func (m *KMSClient) DescribeKey(in *kms.DescribeKeyInput) (*kms.DescribeKeyOutput, error) {
	return &kms.DescribeKeyOutput{
		KeyMetadata: &kms.KeyMetadata{
			Arn: to.Strp("arn:aws:kms:us-east-1:000000000000:key/00000000-0000-0000-0000-000000000000"),
		},
	}, nil
}

func (m *KMSClient) ListResourceTags(in *kms.ListResourceTagsInput) (*kms.ListResourceTagsOutput, error) {
	return &kms.ListResourceTagsOutput{
		Tags: []*kms.Tag{
			&kms.Tag{
				TagKey:   to.Strp("FenrirAllAllowed"),
				TagValue: to.Strp("true"),
			},
		},
	}, nil
}
