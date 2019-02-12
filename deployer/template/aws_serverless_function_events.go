package template

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/utils/to"
	"github.com/grahamjenson/goformation/cloudformation"
	"github.com/grahamjenson/goformation/cloudformation/resources"
)

func ValidateAPIEvent(template *cloudformation.Template, event *resources.AWSServerlessFunction_ApiEvent) error {
	if event == nil {
		return fmt.Errorf("Event Properties nil")
	}

	// Api must be explicitly defined
	// Ensure RestAPI must be nil OR Ref to local RestAPI ID
	if event.RestApiId == "" {
		return fmt.Errorf("RestApiId must be explicitly defined")
	}

	// decode the base64 reference
	ref, err := decodeRef(event.RestApiId)
	if err != nil || ref == "" {
		return fmt.Errorf("RestApiId must be !Ref")
	}

	if _, err := template.GetAWSServerlessApiWithName(ref); err != nil {
		return fmt.Errorf("RestApiId Reference %q not found", ref)
	}

	// Reference must point to API
	return nil
}

func ValidateS3Event(projectName, configName string, event *resources.AWSServerlessFunction_S3Event, s3c aws.S3API) error {
	tags, err := s3.GetBucketTags(s3c, to.Strp(event.Bucket))
	if err != nil {
		return err
	}

	return hasEventTags(projectName, configName, tags)
}

func ValidateKinesisEvent(projectName, configName string, event *resources.AWSServerlessFunction_KinesisEvent, kinc aws.KINAPI) error {
	//event.Stream is an arn e.g. arn:aws:kinesis:us-east-1:000000000000:stream/<stream-name>
	_, _, resource := to.ArnRegionAccountResource(event.Stream)
	// resource is the type/name e.g. "stream/<stream-name>""
	typeName := strings.SplitN(resource, "/", 2)

	if len(typeName) != 2 {
		return fmt.Errorf("Stream incorrect ARN")
	}

	out, err := kinc.ListTagsForStream(&kinesis.ListTagsForStreamInput{
		StreamName: to.Strp(typeName[1]),
	})

	if err != nil {
		return err
	}

	tags := map[string]string{}
	for _, tag := range out.Tags {
		if tag.Key == nil {
			continue
		}
		tags[*tag.Key] = to.Strs(tag.Value)
	}

	return hasEventTags(projectName, configName, tags)
}

func ValidateDynamoDBEvent(projectName, configName string, event *resources.AWSServerlessFunction_DynamoDBEvent, ddbc aws.DDBAPI) error {
	out, err := ddbc.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
		ResourceArn: to.Strp(event.Stream),
	})

	if err != nil {
		return err
	}

	tags := map[string]string{}
	for _, tag := range out.Tags {
		if tag.Key == nil {
			continue
		}
		tags[*tag.Key] = to.Strs(tag.Value)
	}

	return hasEventTags(projectName, configName, tags)
}

func ValidateSQSEvent(projectName, configName string, event *resources.AWSServerlessFunction_SQSEvent, sqsc aws.SQSAPI) error {
	out, err := sqsc.ListQueueTags(&sqs.ListQueueTagsInput{
		QueueUrl: to.Strp(event.Queue),
	})

	if err != nil {
		return err
	}

	tags := map[string]string{}
	for key, value := range out.Tags {
		tags[key] = to.Strs(value)
	}

	return hasEventTags(projectName, configName, tags)
}

func ValidateScheduleEvent(event *resources.AWSServerlessFunction_ScheduleEvent) error {
	// Allowed any
	return nil
}

func ValidateCloudWatchEventEvent(event *resources.AWSServerlessFunction_CloudWatchEventEvent) error {
	// Allowed any
	return nil
}

func hasEventTags(projectName, configName string, tags map[string]string) error {
	if tags["ProjectName"] == projectName && tags["ConfigName"] == configName {
		return nil
	}

	if tags[fmt.Sprintf("FenrirAllowed:%v:%v", projectName, configName)] != "" {
		return nil
	}

	if tags["FenrirAllowed:_all:_all"] != "" {
		return nil
	}

	return fmt.Errorf("ProjectName (%v != %v) OR ConfigName (%v != %v) tags incorrect", tags["ProjectName"], projectName, tags["ConfigName"], configName)
}
