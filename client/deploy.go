package client

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/deployer"
	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/execution"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
)

// Deploy attempts to deploy release
func Deploy(step_fn *string, releaseFile *string) error {
	region, accountID := to.RegionAccount()

	if is.EmptyStr(region) || is.EmptyStr(accountID) {
		return fmt.Errorf("AWS_REGION and AWS_ACCOUNT_ID envars, maybe use assume-role")
	}

	release, err := releaseFromFile(releaseFile, region, accountID)
	if err != nil {
		return err
	}

	deployerARN := to.StepArn(region, accountID, step_fn)

	return deploy(&aws.ClientsStr{}, release, deployerARN, releaseFile)
}

func deploy(awsc aws.Clients, release *deployer.Release, deployerARN *string, releaseFile *string) error {

	release.S3URISHA256s = map[string]string{}

	// replace CodeURI with s3 path to uploaded zip
	// Also write fileSHA
	for name, res := range release.Template.GetAllAWSServerlessFunctionResources() {
		zipPath := zipFilePath(releaseFile, name)
		s3URI := s3FileURI(release, name)

		fileSHA, err := to.SHA256File(zipPath)
		if err != nil {
			return err
		}

		res.CodeUri.String = &s3URI
		release.S3URISHA256s[s3URI] = fileSHA

		err = s3.PutFile(
			awsc.S3(nil, nil, nil),
			to.Strp(zipPath),
			release.Bucket,
			to.Strp(s3FilePath(release, name)),
		)

		if err != nil {
			return err
		}
	}

	// Uploading the Release to S3 to match SHAs
	if err := s3.PutStruct(awsc.S3(nil, nil, nil), release.Bucket, release.ReleasePath(), release); err != nil {
		return err
	}

	exec, err := findOrCreateExec(awsc.SFN(nil, nil, nil), deployerARN, release)
	if err != nil {
		return err
	}

	// Execute every second
	exec.WaitForExecution(awsc.SFN(nil, nil, nil), 1, waiter)

	fmt.Println("")

	if exec.Output == nil {
		return nil
	}

	var outRelease *deployer.Release
	if err := json.Unmarshal([]byte(*exec.Output), &outRelease); err != nil {
		return err
	}

	if outRelease.LogSummary != nil {
		fmt.Println(*outRelease.LogSummary)
	}

	fmt.Println("")
	fmt.Println(to.PrettyJSON(outRelease.Outputs))

	return nil
}

func findOrCreateExec(sfnc sfniface.SFNAPI, deployer *string, release *deployer.Release) (*execution.Execution, error) {
	exec, err := execution.FindExecution(sfnc, deployer, release.ExecutionPrefix())
	if err != nil {
		return nil, err
	}

	if exec != nil {
		return exec, nil
	}

	return execution.StartExecution(sfnc, deployer, release.ExecutionName(), release)
}
