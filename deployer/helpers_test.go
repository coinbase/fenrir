package deployer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/coinbase/fenrir/aws/mocks"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/machine"
	"github.com/coinbase/step/utils/to"
	"github.com/grahamjenson/goformation"
	"github.com/grahamjenson/goformation/cloudformation"
	"github.com/grahamjenson/goformation/intrinsics"
	"github.com/stretchr/testify/assert"
)

////////
// RELEASE
////////
func parseTemplate(rawSAM string) (*cloudformation.Template, error) {
	// process Globals
	// Dont process intrinsics
	return goformation.ParseYAMLWithOptions([]byte(rawSAM), &intrinsics.ProcessorOptions{
		IntrinsicHandlerOverrides: cloudformation.EncoderIntrinsics,
	})
}

func MockRelease(fileName string) (*Release, error) {
	basicSAM, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	template, err := parseTemplate(string(basicSAM))
	if err != nil {
		return nil, err
	}

	release := &Release{
		Release: bifrost.Release{
			AwsAccountID: to.Strp("00000000"),
			ReleaseID:    to.Strp("release-1"),
			ProjectName:  to.Strp("project"),
			ConfigName:   to.Strp("development"),
			CreatedAt:    to.Timep(time.Now()),
		},
		Template: template,
		S3URISHA256s: map[string]string{
			"s3://bucket/path.zip": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	return release, nil
}

func MockAwsClients(release *Release) *mocks.MockClients {

	awsc := mocks.MockAWS()

	raw, _ := json.Marshal(release)
	accountID := release.AwsAccountID
	if accountID == nil {
		accountID = to.Strp("000000000000")
	}

	if release.ProjectName != nil && release.ConfigName != nil && release.ReleaseID != nil {
		releasePath := fmt.Sprintf("%v/%v/%v/%v/release", *accountID, *release.ProjectName, *release.ConfigName, *release.ReleaseID)
		awsc.S3Client.AddGetObject(releasePath, string(raw), nil)

		awsc.S3Client.AddGetObject("path.zip", "", nil)

		// Good resources
		awsc.EC2Client.AddSecurityGroup("sg_correct", *release.ProjectName, *release.ConfigName, "hello", nil)
		awsc.EC2Client.AddSubnet("subnet_correct", "subnet-1", true)
		awsc.IAMClient.AddGetRole("role_correct", *release.ProjectName, *release.ConfigName, "_all")

		// Event Resources
		tags := map[string]string{"ProjectName": "project", "ConfigName": "development"}
		awsc.S3Client.SetBucketTags("bucket", tags, nil)

		// Bad Resources
		awsc.EC2Client.AddSecurityGroup("sg_bad", "bad", *release.ConfigName, "hello", nil)
		awsc.EC2Client.AddSubnet("subnet_bad", "subnet-2", false)
		awsc.IAMClient.AddGetRole("role_bad", "bad", *release.ConfigName, "hello")
	}

	return awsc
}

func createTestStateMachine(t *testing.T, awsc *mocks.MockClients) *machine.StateMachine {
	stateMachine, err := StateMachine()
	assert.NoError(t, err)

	stateMachine.SetTaskFnHandlers(CreateTaskHandlers(awsc))

	assert.NoError(t, err)

	return stateMachine
}

func assertSuccessfulExecution(t *testing.T, release *Release) map[string]interface{} {
	stateMachine := createTestStateMachine(t, MockAwsClients(release))

	exec, err := stateMachine.Execute(release)
	output := exec.Output

	assert.NoError(t, err)
	assert.Equal(t, true, output["success"])
	assert.NotRegexp(t, "error", exec.LastOutputJSON)

	assert.Equal(t, []string{
		"Validate",
		"Lock",
		"CreateChangeSet",
		"WaitForChangeSet",
		"UpdateChangeSet",
		"Execute?",
		"Execute",
		"WaitForComplete",
		"UpdateStack",
		"Complete?",
		"ReleaseLock",
		"Success?",
		"Success",
	}, exec.Path())

	return exec.Output
}

func assertFailedExecution(t *testing.T, release *Release) (map[string]interface{}, []string, string) {
	stateMachine := createTestStateMachine(t, MockAwsClients(release))

	exec, err := stateMachine.Execute(release)

	assert.Error(t, err)
	assert.Regexp(t, "error", exec.LastOutputJSON)

	assert.NotEqual(t, []string{
		"Validate",
		"Lock",
		"CreateChangeSet",
		"WaitForChangeSet",
		"UpdateChangeSet",
		"Execute?",
		"Execute",
		"WaitForComplete",
		"UpdateStack",
		"Complete?",
		"ReleaseLock",
		"Success?",
		"Success",
	}, exec.Path())

	output := exec.LastOutput["error"]
	errorOutput := output.(map[string]interface{})

	return exec.LastOutput, exec.Path(), fmt.Sprintf("%v", errorOutput["Cause"])
}
