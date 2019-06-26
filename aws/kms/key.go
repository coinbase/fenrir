package kms

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

// SecurityGroup struct
type Key struct {
	Id   string
	Tags map[string]string
}

func FindKey(kmsc aws.KMSAPI, keyalias string) (*Key, error) {
	desc, err := kmsc.DescribeKey(&kms.DescribeKeyInput{
		KeyId: to.Strp(keyalias), // Can be alias/key_alias, or ARN or ID
	})

	if err != nil {
		return nil, err
	}

	if desc.KeyMetadata == nil || desc.KeyMetadata.Arn == nil {
		return nil, fmt.Errorf("Cannot find key %q", keyalias)
	}

	arnSplit := strings.SplitN(*desc.KeyMetadata.Arn, "/", 2)

	if len(arnSplit) != 2 {
		return nil, fmt.Errorf("Key incorrect ARN")
	}

	key := Key{
		Id:   arnSplit[1],
		Tags: map[string]string{},
	}

	tagsout, err := kmsc.ListResourceTags(&kms.ListResourceTagsInput{KeyId: desc.KeyMetadata.Arn})

	if err != nil {
		return nil, err
	}

	if tagsout == nil || tagsout.Tags == nil {
		return nil, fmt.Errorf("Cannot find key Tags %q", keyalias)
	}

	for _, tag := range tagsout.Tags {
		if tag.TagKey == nil || tag.TagValue == nil {
			continue
		}
		key.Tags[*tag.TagKey] = *tag.TagValue
	}

	return &key, nil
}
