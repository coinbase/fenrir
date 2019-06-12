package mocks

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

var defaultSnsPolicy = `
{
  "Version": "2008-10-17",
  "Id": "__default_policy_ID",
  "Statement": [
    {
      "Sid": "__default_statement_ID",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "role_correct",
          "*"
        ]
      },
      "Action": "SNS:Subscribe",
      "Resource": "arn:aws:sns:us-east-1:000000000000:test-topic",
      "Condition": {
        "StringEquals": {
          "AWS:SourceOwner": "000000000000"
        }
      }
    }
  ]
}
`

// SNSClient returns
type SNSClient struct {
	aws.SNSAPI
}

// ListTagsForResource returns
func (m *SNSClient) ListTagsForResource(in *sns.ListTagsForResourceInput) (*sns.ListTagsForResourceOutput, error) {
	return &sns.ListTagsForResourceOutput{
		Tags: []*sns.Tag{
			&sns.Tag{
				Key:   to.Strp("ProjectName"),
				Value: to.Strp("project"),
			},
			&sns.Tag{
				Key:   to.Strp("ConfigName"),
				Value: to.Strp("development"),
			},
		},
	}, nil
}
