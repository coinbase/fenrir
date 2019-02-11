package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/coinbase/step/utils/is"
	"github.com/grahamjenson/goformation/intrinsics"

	"github.com/coinbase/fenrir/deployer"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/execution"
	"github.com/coinbase/step/utils/to"

	"github.com/grahamjenson/goformation"
	"github.com/grahamjenson/goformation/cloudformation"
	"github.com/sanathkr/yaml"
)

func prepareRelease(release *deployer.Release, region *string, accountID *string) {
	release.ReleaseID = to.TimeUUID("release-")
	release.CreatedAt = to.Timep(time.Now())

	release.SetDefaults(region, accountID)
}

type ProjectConfig struct {
	ProjectName *string `json:"ProjectName"`
	ConfigName  *string `json:"ConfigName"`
}

func parseRelease(releaseFile string) (*deployer.Release, string, error) {
	rawSAM, err := ioutil.ReadFile(releaseFile)
	if err != nil {
		return nil, "", err
	}

	var projectConfig ProjectConfig
	if err := yaml.Unmarshal(rawSAM, &projectConfig); err != nil {
		return nil, "", err
	}

	release := deployer.Release{}
	release.ProjectName = projectConfig.ProjectName
	release.ConfigName = projectConfig.ConfigName

	if is.EmptyStr(release.ProjectName) || is.EmptyStr(release.ConfigName) {
		return nil, "", fmt.Errorf("ProjectName or ConfigName is nil")
	}

	return &release, string(rawSAM), nil
}

func parseTemplate(rawSAM string) (*cloudformation.Template, error) {
	// process Globals
	// Dont process intrinsics
	template, err := goformation.ParseYAMLWithOptions([]byte(rawSAM), &intrinsics.ProcessorOptions{
		IntrinsicHandlerOverrides: cloudformation.EncoderIntrinsics,
	})

	if err != nil {
		return nil, err
	}

	y, err := template.JSON()
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", string(y))

	// validate
	return template, nil
}

func zipFilePath(releaseFile *string, name string) string {
	return fmt.Sprintf("%v.%v.zip", *releaseFile, name)
}

func s3FilePath(release *deployer.Release, name string) string {
	return fmt.Sprintf("%v/%v.zip", *release.ReleasePath(), name)
}

func s3FileURI(release *deployer.Release, name string) string {
	return fmt.Sprintf("s3://%v/%v", *release.Bucket, s3FilePath(release, name))
}

func prepareTemplate(release *deployer.Release, rawSAM string) string {
	// TODO maybe sub later
	return rawSAM
}

func releaseFromFile(releaseFile *string, region *string, accountID *string) (*deployer.Release, error) {
	release, rawSAM, err := parseRelease(*releaseFile)
	if err != nil {
		return nil, err
	}

	prepareRelease(release, region, accountID)

	rawSAM = prepareTemplate(release, rawSAM)

	// TODO add CodeSHA256
	// Set all Service LambdaSHA values
	template, err := parseTemplate(rawSAM)
	if err != nil {
		return nil, err
	}

	release.Template = template

	if err := release.ValidateSchema(); err != nil {
		return nil, err
	}

	return release, nil
}

func waiter(ed *execution.Execution, sd *execution.StateDetails, err error) error {
	if err != nil {
		return fmt.Errorf("Unexpected Error %v", err.Error())
	}

	var releaseError struct {
		Error *bifrost.ReleaseError `json:"error,omitempty"`
	}

	if sd != nil && sd.LastOutput != nil {
		json.Unmarshal([]byte(*sd.LastOutput), &releaseError)
	}

	fmt.Printf("\rExecution: %v", *ed.Status)
	if releaseError.Error != nil {
		fmt.Printf("\nError: %v\nCause: %v\n", to.Strs(releaseError.Error.Error), to.Strs(releaseError.Error.Cause))
	}

	return nil
}
