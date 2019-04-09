package mocks

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

type DDBClient struct {
	aws.DDBAPI
}

// ListTagsOfResource returns
func (m *DDBClient) ListTagsOfResource(in *dynamodb.ListTagsOfResourceInput) (*dynamodb.ListTagsOfResourceOutput, error) {
	return &dynamodb.ListTagsOfResourceOutput{
		Tags: []*dynamodb.Tag{
			&dynamodb.Tag{Key: to.Strp("ProjectName"), Value: to.Strp("project")},
			&dynamodb.Tag{Key: to.Strp("ConfigName"), Value: to.Strp("development")},
		},
	}, nil
}
