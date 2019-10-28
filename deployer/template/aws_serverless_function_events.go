package template

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/awslabs/goformation/v3/cloudformation"
	"github.com/awslabs/goformation/v3/cloudformation/serverless"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/utils/to"
)

func ValidateAPIEvent(template *cloudformation.Template, event *serverless.Function_ApiEvent) error {
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

	if _, err := template.GetServerlessApiWithName(ref); err != nil {
		return fmt.Errorf("RestApiId Reference %q not found", ref)
	}

	// Reference must point to API
	return nil
}

func ValidateS3Event(projectName, configName string, event *serverless.Function_S3Event, s3c aws.S3API) error {
	tags, err := s3.GetBucketTags(s3c, to.Strp(event.Bucket))
	if err != nil {
		return err
	}

	return hasCorrectTags(projectName, configName, tags)
}

func ValidateKinesisEvent(projectName, configName, region, accountId string, event *serverless.Function_KinesisEvent, kinc aws.KINAPI) error {
	if !strings.HasPrefix(event.Stream, "arn:") {
		event.Stream = fmt.Sprintf("arn:aws:kinesis:%s:%s:%s", region, accountId, event.Stream)
	}

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

	return hasCorrectTags(projectName, configName, tags)
}

func ValidateDynamoDBEvent(projectName, configName string, event *serverless.Function_DynamoDBEvent, ddbc aws.DDBAPI) error {
	// we want to check the tags on the table itself, streams do not have tags
	dynamodbStreamName := strings.SplitN(event.Stream, "/stream", 3)[0]

	out, err := ddbc.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
		ResourceArn: to.Strp(dynamodbStreamName),
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

	return hasCorrectTags(projectName, configName, tags)
}

func ValidateSQSEvent(projectName, configName, region, accountId string, event *serverless.Function_SQSEvent, sqsc aws.SQSAPI) error {
	// If the event is a valid GetAtt
	ref, err := decodeGetAtt(event.Queue)
	if err == nil && len(ref) > 0 {
		return nil
	}

	if !strings.HasPrefix(event.Queue, "arn:") {
		event.Queue = fmt.Sprintf("arn:aws:sqs:%s:%s:%s", region, accountId, event.Queue)
	}

	// event.Queue is ARN e.g. arn:aws:sqs:us-east-1:000000000000:test-queue
	region, account, resource := to.ArnRegionAccountResource(event.Queue)
	if region == "" || account == "" || resource == "" {
		return fmt.Errorf("invalid SQS ARN")
	}

	// need URL e.g. https://sqs.us-east-1.amazonaws.com/000000000000/test-queue
	queueURL := fmt.Sprintf("https://sqs.%v.amazonaws.com/%v/%v", region, accountId, resource)

	out, err := sqsc.ListQueueTags(&sqs.ListQueueTagsInput{
		QueueUrl: &queueURL,
	})

	if err != nil {
		return err
	}

	tags := map[string]string{}
	for key, value := range out.Tags {
		tags[key] = to.Strs(value)
	}

	return hasCorrectTags(projectName, configName, tags)
}

func ValidateSNSEvent(projectName, configName, region, accountId string, event *serverless.Function_SNSEvent, snsc aws.SNSAPI) error {
	// event.Topic is ARN or NAME e.g.arn:aws:sns:us-east-1:000000000000:test-topic
	if strings.HasPrefix(event.Topic, "arn:") {
		region, account, resource := to.ArnRegionAccountResource(event.Topic)
		if region == "" || account == "" || resource == "" {
			return fmt.Errorf("invalid SNS ARN")
		}
	} else {
		event.Topic = fmt.Sprintf("arn:aws:sns:%s:%s:%s", region, accountId, event.Topic)
	}

	out, err := snsc.ListTagsForResource(&sns.ListTagsForResourceInput{
		ResourceArn: &event.Topic,
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

	return hasCorrectTags(projectName, configName, tags)
}

func ValidateScheduleEvent(event *serverless.Function_ScheduleEvent) error {
	// Allowed any
	return nil
}

func ValidateCloudWatchEventEvent(event *serverless.Function_CloudWatchEventEvent) error {
	// Allowed any
	return nil
}
