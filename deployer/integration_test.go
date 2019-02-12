package deployer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

///////////////
// Successful Tests
///////////////

var goodReleases = []string{
	"../examples/tests/allowed/api_private.yml",
	"../examples/tests/allowed/api.yml",
	"../examples/tests/allowed/function.yml",
	"../examples/tests/allowed/function_api.yml",
	"../examples/tests/allowed/layer.yml",
	"../examples/tests/allowed/table.yml",
	"../examples/tests/allowed/s3_event.yml",
	"../examples/tests/allowed/kinesis_event.yml",
	"../examples/tests/allowed/dynamo_event.yml",
}

func Test_Successful_Execution(t *testing.T) {
	for _, releaseFile := range goodReleases {

		t.Run(releaseFile, func(t *testing.T) {
			release, err := MockRelease(releaseFile)

			assert.NoError(t, err)
			assertSuccessfulExecution(t, release)
		})

	}

}

///////////////
// Unsuccessful Tests
///////////////
var badFiles = []struct {
	File     string
	ErrorStr string
}{
	{
		File:     "../examples/tests/not/bad_api_name.yml",
		ErrorStr: "AWS::Serverless::Api#helloAPI: Names are overwritten",
	},
	{
		File:     "../examples/tests/not/bad_function_name.yml",
		ErrorStr: "AWS::Serverless::Function#hello: Names are overwritten",
	},
	{
		File:     "../examples/tests/not/bad_role.yml",
		ErrorStr: `AWS::Serverless::Function#hello: Incorrect ProjectName for Role: has "bad" requires "project"`,
	},
	{
		File:     "../examples/tests/not/bad_security_group.yml",
		ErrorStr: `AWS::Serverless::Function#hello: VpcConfig Incorrect ProjectName for SecurityGroup: has "bad" requires "project"`,
	},
	{
		File:     "../examples/tests/not/bad_subnet.yml",
		ErrorStr: `AWS::Serverless::Function#hello: VpcConfig Validate Subnet Error DeployWithFenrir Tag is nil`,
	},
	{
		File:     "../examples/tests/not/bad_transform.yml",
		ErrorStr: `Transform must be one of the following: "AWS::Serverless-2016-10-31"`,
	},
	{
		File:     "../examples/tests/not/cannot_find_role.yml",
		ErrorStr: `AWS::Serverless::Function#hello: role_unknown not found role err`,
	},
	{
		File:     "../examples/tests/not/cannot_find_security_group.yml",
		ErrorStr: `AWS::Serverless::Function#hello: VpcConfig Find Security Group Error SecurityGroup 'sg_unknown': not found`,
	},
	{
		File:     "../examples/tests/not/cannot_find_subnet.yml",
		ErrorStr: `AWS::Serverless::Function#hello: VpcConfig Find Subnet Error Incorrect Number of Subnets Found. Found 0, Required 1`,
	},
	{
		File:     "../examples/tests/not/external_api_ref.yml",
		ErrorStr: `RestApiId must be !Ref`,
	},
	{
		File:     "../examples/tests/not/invalid_api_ref.yml",
		ErrorStr: `RestApiId Reference "helloAPI" not found`,
	},
	{
		File:     "../examples/tests/not/no_explicit_api.yml",
		ErrorStr: `RestApiId must be explicitly defined`,
	},
	{
		File:     "../examples/tests/not/unsupported_function_event.yml",
		ErrorStr: `Unsupported Event type "IoTRule"`,
	},
	{
		File:     "../examples/tests/not/bad_codeuri_sha.yml",
		ErrorStr: `CodeUri s3://no_sha/path.zip not included in the SHA256s`,
	},
	{
		File:     "../examples/tests/not/invalid_schema.yml",
		ErrorStr: `CodeUri is required`,
	},
	{
		File:     "../examples/tests/not/invalid_event_schema.yml",
		ErrorStr: `StartingPosition is required`,
	},
}

func Test_Unsuccessful_Execution(t *testing.T) {
	for _, test := range badFiles {
		t.Run(test.File, func(t *testing.T) {
			release, err := MockRelease(test.File)
			assert.NoError(t, err)

			_, path, errStr := assertFailedExecution(t, release)
			assert.Equal(t, []string{
				"Validate",
				"FailureClean",
			}, path)

			assert.Regexp(t, test.ErrorStr, errStr)
		})
	}
}
