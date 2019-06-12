package template

import (
	"testing"

	"github.com/awslabs/goformation/cloudformation"
	"github.com/awslabs/goformation/cloudformation/resources"
	"github.com/stretchr/testify/assert"
)

func TestValidateAPIEventWorks(t *testing.T) {
	template, err := MockTemplate("../../examples/tests/allowed/api.yml")
	assert.NoError(t, err)

	err = ValidateAPIEvent(template, &resources.AWSServerlessFunction_ApiEvent{
		RestApiId: cloudformation.Ref("helloAPI"),
	})
	assert.NoError(t, err)
}

func TestValidateS3EventWorks(t *testing.T) {

	awsc := MockAwsClients()
	err := ValidateS3Event("project", "development", &resources.AWSServerlessFunction_S3Event{
		Bucket: "bucket",
	}, awsc.S3(nil, nil, nil))
	assert.NoError(t, err)

}

func TestValidateKinesisEventWorks(t *testing.T) {

	awsc := MockAwsClients()
	err := ValidateKinesisEvent("project", "development", &resources.AWSServerlessFunction_KinesisEvent{
		Stream: "arn:aws:kinesis:us-east-1:000000000000:stream/<stream-name>",
	}, awsc.KIN(nil, nil, nil))
	assert.NoError(t, err)

}

func TestValidateDynamoDBEventWorks(t *testing.T) {

	awsc := MockAwsClients()
	err := ValidateDynamoDBEvent("project", "development", &resources.AWSServerlessFunction_DynamoDBEvent{
		Stream: "db",
	}, awsc.DDB(nil, nil, nil))
	assert.NoError(t, err)

}

func TestValidateSQSEventWorks(t *testing.T) {

	awsc := MockAwsClients()
	err := ValidateSQSEvent("project", "development", &resources.AWSServerlessFunction_SQSEvent{
		Queue: "arn:aws:sqs:us-east-1:000000000000:test-queue",
	}, awsc.SQS(nil, nil, nil))
	assert.NoError(t, err)

}

// SNSEVent

func TestValidateSNSEventWorks(t *testing.T) {

	awsc := MockAwsClients()
	err := ValidateSNSEvent("project", "development", "region", "accountID", &resources.AWSServerlessFunction_SNSEvent{
		Topic: "arn:aws:sns:us-east-1:000000000000:test-topic",
	}, awsc.SNS(nil, nil, nil))
	assert.NoError(t, err)
}

func TestValidateSNSEventWorksWithName(t *testing.T) {

	awsc := MockAwsClients()
	event := resources.AWSServerlessFunction_SNSEvent{
		Topic: "test-topic",
	}
	err := ValidateSNSEvent("project", "development", "region", "accountID", &event, awsc.SNS(nil, nil, nil))
	assert.NoError(t, err)
	assert.Equal(t, "arn:aws:sns:region:accountID:test-topic", event.Topic)
}

func TestValidateSNSEventDoesntWorkWithIncorrectTags(t *testing.T) {

	awsc := MockAwsClients()
	err := ValidateSNSEvent("project", "wrong_config", "region", "accountID", &resources.AWSServerlessFunction_SNSEvent{
		Topic: "arn:aws:sns:us-east-1:000000000000:test-topic",
	}, awsc.SNS(nil, nil, nil))
	assert.Error(t, err)
}

// ScheduleEvent

func TestValidateScheduleEventWorks(t *testing.T) {
	err := ValidateScheduleEvent(&resources.AWSServerlessFunction_ScheduleEvent{})
	assert.NoError(t, err)
}

func TestValidateCloudWatchEventEventWorks(t *testing.T) {

	err := ValidateCloudWatchEventEvent(&resources.AWSServerlessFunction_CloudWatchEventEvent{})
	assert.NoError(t, err)

}
