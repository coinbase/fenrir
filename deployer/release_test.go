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
