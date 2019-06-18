package static

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

func StaticSiteResources(awsc aws.Clients) cfn.CustomResourceLambdaFunction {
	// Wrapper adds the call backs to Cloudformation
	return cfn.LambdaWrap(staticSiteResources(awsc))
}

func staticSiteResources(awsc aws.Clients) func(context.Context, cfn.Event) (string, map[string]interface{}, error) {
	return func(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
		v, _ := event.ResourceProperties["Echo"].(string)

		data := map[string]interface{}{
			"Echo": v,
		}

		fmt.Println(to.PrettyJSON(event))
		fmt.Println("")
		fmt.Println(to.PrettyJSON(data))

		return "asd", data, nil
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
              "Custom::S3File"
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
