package custom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/deployer/template"
	"github.com/coinbase/step/utils/to"
)

var assumedRole = to.Strp("coinbase-fenrir-assumed")

type SNSEvent struct {
	raw     string
	Records []SNSEventRecord `json:"Records"`
}

func (message *SNSEvent) UnmarshalJSON(data []byte) error {
	type xSNSEvent SNSEvent
	var rawMessageX xSNSEvent

	if err := json.Unmarshal(data, &rawMessageX); err != nil {
		return err
	}

	*message = SNSEvent{
		Records: rawMessageX.Records,
		raw:     string(data),
	}

	return nil
}

type SNSEventRecord struct {
	SNS SNSEntity `json:"Sns"`
}

type SNSEntity struct {
	Message string `json:"Message"`
}

// Wrap the
func CustomResourceFn(awsc aws.Clients) func(context.Context, SNSEvent) (interface{}, error) {
	// SNS fn, wrapping Cloudformation function, wrapping custom resource fn
	return snsWrapperFn(cfn.LambdaWrap(customResourceFn(awsc)))
}

func snsWrapperFn(wrappedFn cfn.CustomResourceLambdaFunction) func(context.Context, SNSEvent) (interface{}, error) {
	return func(ctx context.Context, event SNSEvent) (interface{}, error) {
		// Print out the RawInput for debugging and error handling
		fmt.Println(event.raw)

		if len(event.Records) != 1 {
			return nil, fmt.Errorf("Must have exactly 1 record")
		}

		record := event.Records[0]

		var cfnEvent cfn.Event
		err := json.Unmarshal([]byte(record.SNS.Message), &cfnEvent)

		if err != nil {
			return nil, err
		}

		return wrappedFn(ctx, cfnEvent)
	}
}

func customResourceFn(awsc aws.Clients) func(context.Context, cfn.Event) (string, map[string]interface{}, error) {
	return func(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
		region, accountID, _ := to.ArnRegionAccountResource(event.StackID)
		lambdas3c := awsc.S3(nil, nil, nil)
		s3c := awsc.S3(&region, &accountID, assumedRole)

		// Only Handle Creates and Updates
		if event.RequestType == cfn.RequestDelete {
			return event.PhysicalResourceID, map[string]interface{}{}, nil
		}

		switch event.ResourceType {
		case "Custom::S3File":
			bucket, key, uri, err := template.ValidateS3FilePropertyValues(event.ResourceProperties["Bucket"], event.ResourceProperties["Key"], event.ResourceProperties["Uri"])
			if err != nil {
				return "", nil, err
			}
			return handleS3File(bucket, key, uri, lambdas3c, s3c)
		case "Custom::S3ZipFile":
			bucket, key, uri, err := template.ValidateS3FilePropertyValues(event.ResourceProperties["Bucket"], event.ResourceProperties["Key"], event.ResourceProperties["Uri"])
			if err != nil {
				return "", nil, err
			}
			return handleS3ZipFile(bucket, key, uri, lambdas3c, s3c)
		}

		return event.PhysicalResourceID, map[string]interface{}{}, fmt.Errorf("Unknown Resource Type %v", event.ResourceType)
	}
}

func s3UriToBucketKey(uri string) (string, string, error) {
	s3BucketPath := strings.SplitN(strings.TrimPrefix(uri, "s3://"), "/", 2)
	if len(s3BucketPath) != 2 {
		return "", "", fmt.Errorf("Uri incorrect")
	}
	return s3BucketPath[0], s3BucketPath[1], nil
}

// UTILS
func s3Uri(bucket, key string) string {
	return fmt.Sprintf("s3://%v/%v", bucket, key)
}

// http.DetectContentType won't perfectly detect every content type.
// This function provides a place to override content type by looking at the filename:
// * svg: Always set to be an 'image/svg+xml'
func detectContentType(fileName string, content []byte) string {
	extension := filepath.Ext(fileName)
	if extension == ".svg" {
		return "image/svg+xml"
	} else if extension == ".css" {
		return "text/css"
	} else {
		return http.DetectContentType(content)
	}
}

var S3FileSchema = `{
  "additionalProperties": false,
  "properties": {
      "Properties": {
          "additionalProperties": false,
          "properties": {
              "Bucket": { "type": "string" },
              "Key": { "type": "string" },
              "Uri": { "type": "string" }
          },
          "required": ["Bucket", "Key", "Uri"],
          "type": "object"
      },
      "Type": {
          "enum": [
              "Custom::S3File",
              "Custom::S3ZipFile"
          ],
          "type": "string"
      }
  },
  "required": [
      "Type",
      "Properties"
  ],
  "type": "object"
}
`
