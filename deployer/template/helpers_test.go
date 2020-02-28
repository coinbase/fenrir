package template

import (
	"io/ioutil"

	goformation "github.com/awslabs/goformation/v4"
	"github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/intrinsics"
	"github.com/coinbase/fenrir/aws/mocks"
)

////////
// RELEASE
////////

func MockTemplate(fileName string) (*cloudformation.Template, error) {
	basicSAM, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	template, err := goformation.ParseYAMLWithOptions([]byte(string(basicSAM)), &intrinsics.ProcessorOptions{
		IntrinsicHandlerOverrides: cloudformation.EncoderIntrinsics,
	})

	if err != nil {
		return nil, err
	}

	return template, nil
}

func MockS3SHA() string {
	return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
}

func MockAwsClients() *mocks.MockClients {

	awsc := mocks.MockAWS()

	awsc.S3Client.AddGetObject("path.zip", "", nil)

	// Good resources
	awsc.EC2Client.AddSecurityGroup("sg_correct", "project", "development", "rn", nil)
	awsc.EC2Client.AddSubnet("subnet_correct", "subnet-1", true)
	awsc.IAMClient.AddGetRole("role_correct", "project", "development", "_all")

	// Event Resources
	tags := map[string]string{"ProjectName": "project", "ConfigName": "development"}
	awsc.S3Client.SetBucketTags("bucket", tags, nil)

	// Bad Resources
	awsc.EC2Client.AddSecurityGroup("sg_bad", "bad", "development", "rn", nil)
	awsc.EC2Client.AddSubnet("subnet_bad", "subnet-2", false)
	awsc.IAMClient.AddGetRole("role_bad", "bad", "development", "rn")

	return awsc
}
