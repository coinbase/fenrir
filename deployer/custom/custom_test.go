package custom

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/coinbase/fenrir/aws/mocks"
	"github.com/stretchr/testify/assert"
)

// To create this file, create the same directory structure
// Make sure to create these files using echo -n to prevent saving a newline:
//  `echo -n 'I AM ROOT FILE' > root`
//  `echo -n 'I AM FOLDER FILE' > folder/file`
//  `echo -n 'I AM NOT SVG' > folder/test.svg`
//  `echo -n '<svg></svg>' > folder/test2.svg`
//  `echo -n 'p { backround: black; }' > folder/test.css`
// To zip it:
//  `zip new.zip root folder/file folder/test.svg folder/test2.svg folder/test.css`
// To base64 encode it:
//  `base64 new.zip`
var SimpleZIPFile, _ = base64.StdEncoding.DecodeString(
	"UEsDBAoAAAAAAC+JSE5YS6enDgAAAA4AAAAEABwAcm9vdFVUCQADyrddXCi4XVx1eAsAAQT1AQAABBQAAABJIEFNIFJPT1QgRklMRVBLAwQKAAAAAADTiEhOXtqaYBAAAAAQAAAACwAcAGZvbGRlci9maWxlVVQJAAMdt11cH7ddXHV4CwABBPUBAAAEFAAAAEkgQU0gRk9MREVSIEZJTEVQSwMECgAAAAAA2ohITuHgaBoMAAAADAAAAA8AHABmb2xkZXIvdGVzdC5zdmdVVAkAAyu3XVwtt11cdXgLAAEE9QEAAAQUAAAASSBBTSBOT1QgU1ZHUEsDBBQAAAAIAOKISE4GH0AHCgAAAAsAAAAQABwAZm9sZGVyL3Rlc3QyLnN2Z1VUCQADOLddXCi4XVx1eAsAAQT1AQAABBQAAACzKS5Lt7PRB5EAUEsDBAoAAAAAAFeJSE4amjZ3FwAAABcAAAAPABwAZm9sZGVyL3Rlc3QuY3NzVVQJAAMVuF1cFbhdXHV4CwABBPUBAAAEFAAAAHAgeyBiYWNrcm91bmQ6IGJsYWNrOyB9UEsBAh4DCgAAAAAAL4lITlhLp6cOAAAADgAAAAQAGAAAAAAAAQAAAKSBAAAAAHJvb3RVVAUAA8q3XVx1eAsAAQT1AQAABBQAAABQSwECHgMKAAAAAADTiEhOXtqaYBAAAAAQAAAACwAYAAAAAAABAAAApIFMAAAAZm9sZGVyL2ZpbGVVVAUAAx23XVx1eAsAAQT1AQAABBQAAABQSwECHgMKAAAAAADaiEhO4eBoGgwAAAAMAAAADwAYAAAAAAABAAAApIGhAAAAZm9sZGVyL3Rlc3Quc3ZnVVQFAAMrt11cdXgLAAEE9QEAAAQUAAAAUEsBAh4DFAAAAAgA4ohITgYfQAcKAAAACwAAABAAGAAAAAAAAQAAAKSB9gAAAGZvbGRlci90ZXN0Mi5zdmdVVAUAAzi3XVx1eAsAAQT1AQAABBQAAABQSwECHgMKAAAAAABXiUhOGpo2dxcAAAAXAAAADwAYAAAAAAABAAAApIFKAQAAZm9sZGVyL3Rlc3QuY3NzVVQFAAMVuF1cdXgLAAEE9QEAAAQUAAAAUEsFBgAAAAAFAAUAmwEAAKoBAAAAAA==",
)

var TestMessage string = `{
				  "LogicalResourceId": "customFile",
				  "PhysicalResourceId": "physical-id",
				  "RequestId": "req-id",
				  "RequestType": "Create",
				  "ResourceProperties": {
				    "Bucket": "bucket-here",
				    "Key": "index.html",
				    "ServiceToken": "token",
				    "Uri": "s3://uri-id/index.html"
				  },
				  "ResourceType": "Custom::S3File",
				  "ResponseURL": "url",
				  "ServiceToken": "token",
				  "StackId": "stack-if"
				}`
var TestEvent string = `{
  "Records": [
    {
      "EventVersion": "1.0",
      "EventSubscriptionArn": "arn:aws:sns:us-east-1:arn",
      "EventSource": "aws:sns",
      "Sns": {
        "MessageId": "6ae3f7a1-2772-568c-9175-a603bc40bf03",
        "Message": %q
      }
    }
  ]
}`

func Test_SNS_Wrapper(t *testing.T) {
	called := false

	event := fmt.Sprintf(TestEvent, TestMessage)

	testFn := func(_ context.Context, _ cfn.Event) (string, error) {
		called = true
		return "", nil
	}
	fn := snsWrapperFn(testFn)

	var x SNSEvent
	err := json.Unmarshal([]byte(event), &x)
	assert.NoError(t, err)

	value, err := fn(nil, x)

	assert.NoError(t, err)
	assert.Equal(t, "", value)
	assert.True(t, called)
}

func Test_SNS_Wrapper_With_Bad_Message(t *testing.T) {
	called := false

	event := fmt.Sprintf(TestEvent, "bad_message")

	testFn := func(_ context.Context, _ cfn.Event) (string, error) {
		called = true
		return "", nil
	}
	fn := snsWrapperFn(testFn)

	var x SNSEvent
	err := json.Unmarshal([]byte(event), &x)
	assert.NoError(t, err)

	_, err = fn(nil, x)

	assert.Error(t, err)
	assert.False(t, called)
}

func Test_SNS_Wrapper_With_Bad_Event(t *testing.T) {
	called := false

	event := "{}"

	testFn := func(_ context.Context, _ cfn.Event) (string, error) {
		called = true
		return "", nil
	}

	fn := snsWrapperFn(testFn)

	var x SNSEvent
	err := json.Unmarshal([]byte(event), &x)
	assert.NoError(t, err)

	_, err = fn(nil, x)

	assert.Error(t, err)
	assert.False(t, called)
}

func Test_Custom_S3File(t *testing.T) {
	awsc := mocks.MockAWS()
	awsc.S3Client.AddGetObject("fromKey", "", nil)

	fn := customResourceFn(awsc)
	id, data, err := fn(nil, cfn.Event{
		ResourceType: "Custom::S3File",
		RequestType:  "Create",
		StackID:      "arn:aws:cloudformation:us-east-1:00000000000:stack/stack/id",
		ResponseURL:  "http://localhost:8080",
		ResourceProperties: map[string]interface{}{
			"Bucket": "toBucket",
			"Key":    "toKey",
			"Uri":    "s3://fromBucket/fromKey",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, id, "toBucket/toKey")
	assert.Equal(t, data["Uri"], "s3://toBucket/toKey")
}

func Test_Custom_S3ZipFile(t *testing.T) {
	awsc := mocks.MockAWS()
	awsc.S3Client.AddGetObject("fromKey", string(SimpleZIPFile), nil)

	fn := customResourceFn(awsc)
	id, data, err := fn(nil, cfn.Event{
		ResourceType: "Custom::S3ZipFile",
		RequestType:  "Create",
		StackID:      "arn:aws:cloudformation:us-east-1:00000000000:stack/stack/id",
		ResponseURL:  "http://localhost:8080",
		ResourceProperties: map[string]interface{}{
			"Bucket": "toBucket",
			"Key":    "toKey",
			"Uri":    "s3://fromBucket/fromKey",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, id, "toBucket/toKey")
	assert.Equal(t, len(data["Uris"].([]string)), 5)
}
