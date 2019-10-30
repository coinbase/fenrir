package deployer

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_Release_Cleanup(t *testing.T) {
	release, err := MockRelease("../examples/tests/allowed/function.yml")
	assert.NoError(t, err)

	release.ChangeSetType = to.Strp("CREATE")

	awsc := MockAwsClients(release)

	awsc.CFClient.StackResp = &cloudformation.DescribeStacksOutput{Stacks: []*cloudformation.Stack{
		&cloudformation.Stack{
			StackStatus:  to.Strp("ROLLBACK_COMPLETE"),
			CreationTime: to.Timep(time.Now()),
		},
	}}

	err = release.CleanUp(awsc.S3(nil, nil, nil), awsc.CF(nil, nil, nil))
	assert.NoError(t, err)

	assert.True(t, awsc.CFClient.DeleteStackCalled)
}

func Test_Release_CreateChangeSetInput_Tags(t *testing.T) {
	t.Run("tags", func(t *testing.T) {
		release, err := MockRelease("../examples/tests/allowed/function.yml")
		release.ChangeSetTags = map[string]string{
			"CustomTag":   "TTTAG",
			"ProjectName": "SHOULD_NOT_OVERRIDE",
		}
		release.Env = "test-env"
		release.SetDefaults(to.Strp("region"), to.Strp("account"))
		assert.NoError(t, err)

		input, err := release.CreateChangeSetInput()
		assert.NoError(t, err)

		tags := make(map[string]string)
		for _, t := range input.Tags {
			tags[*t.Key] = *t.Value
		}

		assert.Equal(t, "test-env", tags["Env"])
		assert.Equal(t, "project", tags["ProjectName"])
		assert.Equal(t, "development", tags["ConfigName"])
		assert.Equal(t, "TTTAG", tags["CustomTag"])
	})
}
