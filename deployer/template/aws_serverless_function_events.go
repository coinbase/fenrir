package template

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/utils/to"
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

	return hasEventTags(projectName, configName, tags)
}

func ValidateSQSEvent(projectName, configName string, event *resources.AWSServerlessFunction_SQSEvent, sqsc aws.SQSAPI) error {
	// event.Queue is ARN e.g. arn:aws:sqs:us-east-1:000000000000:test-queue
	region, account, resource := to.ArnRegionAccountResource(event.Queue)
	if region == "" || account == "" || resource == "" {
		return fmt.Errorf("invalid SQS ARN")
	}

	// need URL e.g. https://sqs.us-east-1.amazonaws.com/000000000000/test-queue
	queueURL := fmt.Sprintf("https://sqs.%v.amazonaws.com/%v/%v", region, account, resource)

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

	return hasEventTags(projectName, configName, tags)
}

type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

type StatementEntry struct {
	Effect    string
	Action    []string
	Resource  string
	Principal PrincipalEntry
}

type PrincipalEntry struct {
	AWS interface{}
}

func ValidateSNSEvent(projectName, configName string, roleArn string, event *resources.AWSServerlessFunction_SNSEvent, snsc aws.SNSAPI) error {
	// event.Topic is ARN e.g. arn:aws:sns:us-east-1:000000000000:test-topic
	region, account, resource := to.ArnRegionAccountResource(event.Topic)
	if region == "" || account == "" || resource == "" {
		return fmt.Errorf("invalid SNS ARN")
	}

	out, err := snsc.GetTopicAttributes(&sns.GetTopicAttributesInput{
		TopicArn: &event.Topic,
	})
	if err != nil {
		return err
	}

	var policy PolicyDocument
	err = json.Unmarshal([]byte(*out.Attributes["Policy"]), &policy)
	if err != nil {
		return err
	}

	// We want to allow subscriptions for any lambda with the specified role, as long
	// as the role is specified in the topic's access policy principals.
	for _, entry := range policy.Statement {
		valid := false
		for _, action := range entry.Action {
			if action == "sns:Subscribe" {
				valid = true
			}
		}

		if !valid {
			continue
		}

		var principals []string

		switch entry.Principal.AWS.(type) {
		case string:
			principals = []string{entry.Principal.AWS.(string)}
		case []string:
			principals = entry.Principal.AWS.([]string)
		}

		for _, principal := range principals {
			if roleArn == principal {
				return nil
			}
		}
	}

	return fmt.Errorf("sns access policy does not include principal: %s", roleArn)
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

	if tags[fmt.Sprintf("FenrirAllowed:%v:%v", projectName, configName)] == "true" {
		return nil
	}

	if tags["FenrirAllAllowed"] == "true" {
		return nil
	}

	return fmt.Errorf("ProjectName (%v != %v) OR ConfigName (%v != %v) tags incorrect", tags["ProjectName"], projectName, tags["ConfigName"], configName)
}
