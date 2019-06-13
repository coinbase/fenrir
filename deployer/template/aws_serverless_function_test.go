package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAWSServerlessFunctionWorks(t *testing.T) {
	template, err := MockTemplate("../../examples/tests/allowed/function.yml")
	assert.NoError(t, err)

	awsc := MockAwsClients()

	fn, err := template.GetAWSServerlessFunctionWithName("hello")
	assert.NoError(t, err)

	err = ValidateAWSServerlessFunction(
		"project", "development", "region", "account", "rn",
		template,
		fn,
		map[string]string{
			"s3://bucket/path.zip": MockS3SHA(),
		},
		awsc.IAM(nil, nil, nil),
		awsc.EC2(nil, nil, nil),
		awsc.S3(nil, nil, nil),
		awsc.KIN(nil, nil, nil),
		awsc.DDB(nil, nil, nil),
		awsc.SQS(nil, nil, nil),
		awsc.SNS(nil, nil, nil),
		awsc.KMS(nil, nil, nil),
	)

	assert.NoError(t, err)
}
