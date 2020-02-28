package cwl

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/coinbase/fenrir/aws"
)

// HasStack Hack from https://github.com/aws/aws-cli/blob/master/awscli/customizations/cloudformation/deployer.py#L38
func ListLogGroupTags(cwlc aws.CWLAPI, name *string) (map[string]string, error) {
	ret, err := cwlc.ListTagsLogGroup(&cloudwatchlogs.ListTagsLogGroupInput{
		LogGroupName: name,
	})

	if err != nil {
		return nil, err
	}

	tags := map[string]string{}

	for key, value := range ret.Tags {
		if value != nil {
			tags[key] = *value
		}
	}

	return tags, nil
}
