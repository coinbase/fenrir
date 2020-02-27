package deployer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/aws/cf"
	"github.com/coinbase/fenrir/deployer/template"

	"github.com/coinbase/step/aws/s3"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
	"github.com/xeipuuv/gojsonschema"

	gocf "github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/schema"
)

type Release struct {
	bifrost.Release

	// The SAM YAML template
	Template *gocf.Template `json:"template"`

	// All references to S3 must come with SHA values
	S3URISHA256s map[string]string `json:"s3_uris_sha256s,omitempty"`

	StackName *string `json:"stack_name,omitempty"`

	ChangeSetName *string `json:"change_set_name,omitempty"`
	ChangeSetType *string `json:"change_set_type,omitempty"` // CREATE || UPDATE

	// ChangeSetStatus enum CREATE_IN_PROGRESS, CREATE_COMPLETE, or FAILED
	ChangeSetStatus       string `json:"change_set_status,omitempty"`
	ChangeSetStatusReason string `json:"change_set_status_reason,omitempty"`

	// ChangeSetExecutionStatus enum AVAILABLE, UNAVAILABLE, OBSOLETE
	ChangeSetExecutionStatus string `json:"change_set_execution_status,omitempty"`

	// StackStatus enum CREATE_COMPLETE, CREATE_FAILED, CREATE_IN_PROGRESS, DELETE_COMPLETE, DELETE_FAILED, DELETE_IN_PROGRESS, REVIEW_IN_PROGRESS, ROLLBACK_COMPLETE, ROLLBACK_FAILED, ROLLBACK_IN_PROGRESS, UPDATE_COMPLETE, UPDATE_COMPLETE_CLEANUP_IN_PROGRESS, UPDATE_IN_PROGRESS, UPDATE_ROLLBACK_COMPLETE, UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS, UPDATE_ROLLBACK_FAILED, UPDATE_ROLLBACK_IN_PROGRESS
	// or empty_string as it is used in a choice block
	StackStatus string `json:"stack_status"`

	StackCreationTime *time.Time `json:"stack_creation_time,omitempty"` // Can be nil
	StackStatusReason string     `json:"stack_status_reason,omitempty"`

	LogSummary *string `json:"log_summary,omitempty"`

	Outputs map[string]string `json:"outputs,omitempty"`

	// DEPRECATED USE change_set_tags
	Env string `json:"env,omitempty"`

	ChangeSetTags map[string]string `json:"change_set_tags,omitempty"`
}

//////////
// Validate
//////////

// Validate returns
func (release *Release) Validate(s3c aws.S3API) error {
	if err := release.Release.Validate(s3c, &Release{}); err != nil {
		return err
	}

	if release.Template == nil {
		return fmt.Errorf("SAM is nil")
	}

	input, err := release.CreateChangeSetInput()
	if err != nil {
		return err
	}

	// Validate all SHAs are correct
	if err := release.ValidateSHAs(s3c); err != nil {
		return err
	}

	if err := input.Validate(); err != nil {
		return err
	}

	if err := release.ValidateSchema(); err != nil {
		return err
	}

	return nil
}

func (release *Release) ValidateSHAs(s3c aws.S3API) error {
	for s3URL, sha := range release.S3URISHA256s {
		// urls starts with s3://<bucket>/<path>
		if !strings.HasPrefix(s3URL, "s3://") {
			return fmt.Errorf("S3 URL in S3URISHA256s does not start with s3://")
		}

		s3BucketPath := strings.SplitN(strings.TrimPrefix(s3URL, "s3://"), "/", 2)
		if len(s3BucketPath) != 2 {
			return fmt.Errorf("S3 URL incorrect")
		}

		s3SHA, err := s3.GetSHA256(s3c, to.Strp(s3BucketPath[0]), to.Strp(s3BucketPath[1]))
		if err != nil {
			return err
		}

		if s3SHA != sha {
			return fmt.Errorf("Incorrect SHA for %v: is %v expected %v", s3URL, s3SHA, sha)
		}
	}
	return nil
}

func (release *Release) ValidateSchema() error {
	// Don't use SAM.JSON() because it replaces base64 strings with objects
	templateBody, err := json.Marshal(release.Template)
	if err != nil {
		return err
	}

	schemaLoader := gojsonschema.NewStringLoader(schema.SamSchema)
	documentLoader := gojsonschema.NewStringLoader(string(templateBody))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		errors := ""
		errors += "The document is not valid. see errors :\n"
		for _, desc := range result.Errors() {
			errors += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf(errors)
	}

	return nil
}

// Resource Validations
func (release *Release) ValidateTemplate(
	ec2c aws.EC2API,
	iamc aws.IAMAPI,
	s3c aws.S3API,
	kinc aws.KINAPI,
	ddbc aws.DDBAPI,
	sqsc aws.SQSAPI,
	snsc aws.SNSAPI,
	kmsc aws.KMSAPI,
	lambdac aws.LambdaAPI,
) error {
	// Disabling some template objects because their interations might be
	if release.Template.Parameters != nil {
		return fmt.Errorf("Unsupported Parameters")
	}

	if release.Template.Conditions != nil {
		return fmt.Errorf("Unsupported Conditions")
	}

	if release.Template.Mappings != nil {
		return fmt.Errorf("Unsupported Mappings")
	}

	if release.Template.Metadata != nil {
		return fmt.Errorf("Unsupported Metadata")
	}

	if err := template.ValidateTemplateResources(
		*release.ProjectName, *release.ConfigName,
		*release.AwsRegion, *release.AwsAccountID,
		release.Template, release.S3URISHA256s,
		iamc, ec2c, s3c, kinc, ddbc, sqsc, snsc, kmsc, lambdac); err != nil {
		return err
	}

	return nil
}

//////////
// Defaults
//////////

func (release *Release) SetDefaults(region *string, account *string) {
	release.Success = to.Boolp(false)

	if release.Timeout == nil {
		release.Timeout = to.Intp(300) // Default to 5 mins
	}

	release.Release.SetDefaults(region, account, "coinbase-fenrir-")

	release.StackName = release.CreateStackName() // HARD CODED
	release.ChangeSetName = to.Strp("changeset" + time.Now().Format("20060102T150405Z0700"))

	if release.S3URISHA256s == nil {
		release.S3URISHA256s = map[string]string{}
	}

	// Override Tags
	if release.ChangeSetTags == nil {
		release.ChangeSetTags = map[string]string{}
	}

	release.ChangeSetTags["ProjectName"] = to.Strs(release.ProjectName)
	release.ChangeSetTags["ConfigName"] = to.Strs(release.ConfigName)
	release.ChangeSetTags["ReleaseID"] = to.Strs(release.ReleaseID)

	if release.Env != "" {
		release.ChangeSetTags["Env"] = release.Env
	}
}

func (release *Release) CreateStackName() *string {
	// Must staisfy [a-zA-Z][-a-zA-Z0-9]*
	name := fmt.Sprintf("sam-%v-%v", *release.ProjectName, *release.ConfigName)
	name = strings.Replace(name, "/", "-", -1)
	name = strings.Replace(name, "_", "-", -1)

	return to.Strp(name)
}

//////////
// Deploy/Halt/Health
//////////

// Deploy deploys code
// basically reimplementing this https://docs.aws.amazon.com/cli/latest/reference/cloudformation/deploy/index.html
func (release *Release) CreateChangeSet(cfc aws.CFAPI) error {
	cst, err := cf.ChangeSetType(cfc, release.StackName)

	if err != nil {
		return err
	}

	release.ChangeSetType = cst

	input, err := release.CreateChangeSetInput()
	if err != nil {
		return err
	}

	return cf.CreateChangeSet(cfc, input)
}

func (release *Release) CreateChangeSetInput() (*cloudformation.CreateChangeSetInput, error) {
	templateBody, err := release.Template.JSON()
	if err != nil {
		return nil, err
	}

	changeSetInput := &cloudformation.CreateChangeSetInput{
		ChangeSetName: release.ChangeSetName,
		ClientToken:   release.ReleaseID,
		Description:   to.Strp("Fenrir deploy"),
		StackName:     release.StackName,
		ChangeSetType: release.ChangeSetType,
		Capabilities:  []*string{to.Strp("CAPABILITY_IAM")},
		TemplateBody:  to.Strp(string(templateBody)),
		Tags:          mapToTags(release.ChangeSetTags),
	}

	return changeSetInput, nil
}

func mapToTags(tags map[string]string) []*cloudformation.Tag {
	cstags := []*cloudformation.Tag{}
	for k, v := range tags {
		cstags = append(cstags, &cloudformation.Tag{Key: to.Strp(k), Value: to.Strp(v)})
	}

	return cstags
}

// FetchChangeSet returns two errors (normal error, halt error)
func (release *Release) FetchChangeSet(cfc aws.CFAPI) error {
	// Once a changeset is executed and completed it gets deleted or something
	if release.ChangeSetStatus == "CREATE_COMPLETE" || release.ChangeSetStatus == "FAILED" {
		// No need to further update
		return nil
	}

	output, err := cfc.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
		ChangeSetName: release.ChangeSetName,
		StackName:     release.StackName,
	})

	if err != nil {
		return err
	}

	if output == nil {
		return fmt.Errorf("Unknown DescribeChangeSet Error")
	}

	if output.Status != nil {
		release.ChangeSetStatus = *output.Status
	}

	if output.ExecutionStatus != nil {
		release.ChangeSetExecutionStatus = *output.ExecutionStatus
	}

	if output.StatusReason != nil {
		release.ChangeSetStatusReason = *output.StatusReason
	}

	return nil
}

func (release *Release) ClientRequestToken() *string {
	return release.ChangeSetName
}

func (release *Release) FetchStack(s3c aws.S3API, cfc aws.CFAPI) error {
	stack, err := cf.DescribeStack(cfc, release.StackName)

	if err != nil {
		switch err.(type) {
		case cf.NotFoundError:
			return nil // Not found dont need to delete
		default:
			return err
		}
	}

	return release.updateStack(s3c, cfc, stack)
}

func (release *Release) updateStack(s3c aws.S3API, cfc aws.CFAPI, stack *cloudformation.Stack) error {
	if stack.StackStatus != nil {
		release.StackStatus = *stack.StackStatus
	}

	release.StackCreationTime = stack.CreationTime

	if stack.Outputs != nil {
		release.Outputs = map[string]string{}
		for _, op := range stack.Outputs {
			if op.OutputKey == nil || op.OutputValue == nil {
				continue
			}
			release.Outputs[*op.OutputKey] = *op.OutputValue
		}
	}

	if stack.StackStatusReason != nil {
		release.StackStatusReason = *stack.StackStatusReason
	}

	output, err := cfc.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{StackName: release.StackName})
	if err != nil || output == nil || output.StackEvents == nil {
		return nil // Ignore this error, not great but it will be fine, I swear.
	}

	// LOG looks like
	// date	Status	Type	Logical ID	Status Reason
	log := ""
	for _, e := range output.StackEvents {
		// Filter by token should only show changeset events
		if e.ClientRequestToken != nil && (*e.ClientRequestToken != *release.ClientRequestToken()) {
			continue
		}

		log += fmt.Sprintf(
			"%s %s %s %s %s\n",
			e.Timestamp.Format(time.RFC3339),
			to.Strs(e.ResourceStatus),
			to.Strs(e.ResourceType),
			to.Strs(e.LogicalResourceId),
			to.Strs(e.ResourceStatusReason),
		)
	}

	if release.ChangeSetStatus == "FAILED" && release.ChangeSetStatusReason != "" {
		log += fmt.Sprintf("%s\n", release.ChangeSetStatusReason)
	}

	// Attach log to release and write to file
	release.LogSummary = to.Strp(log)
	release.WriteLog(s3c, log) // ignore errors

	return nil
}

// Execute executes changeset
func (release *Release) Execute(cfc aws.CFAPI) error {
	_, err := cfc.ExecuteChangeSet(&cloudformation.ExecuteChangeSetInput{
		ChangeSetName:      release.ChangeSetName,
		ClientRequestToken: release.ClientRequestToken(),
		StackName:          release.StackName,
	})

	if err != nil {
		return err
	}

	return nil
}

// CleanUpStuckStack checks to see if we need to delete the stack on create failure
// We have to be very careful in this method as we DO NOT want to accidentally delete a stack
// because of https://github.com/awslabs/aws-cdk/issues/901
func (release *Release) CleanUp(s3c aws.S3API, cfc aws.CFAPI) error {
	if release.ChangeSetType == nil || *release.ChangeSetType != "CREATE" {
		return nil
	}

	// Make sure the stack exists and update the
	stack, err := cf.DescribeStack(cfc, release.StackName)

	if err != nil {
		switch err.(type) {
		case cf.NotFoundError:
			return nil // Doesn't exist no neet to delete
		default:
			return err
		}
	}

	if err := release.updateStack(s3c, cfc, stack); err != nil {
		return err
	}

	if release.StackCreationTime == nil {
		return nil // Extra paranoid
	}

	// This is to doubly make sure we are good to delete
	// Was this stack created in the last Half Hour (wiggle 5 mins for time variance)
	if !is.WithinTimeFrame(release.StackCreationTime, 30*time.Minute, 5*time.Minute) {
		return nil
	}

	if !(release.ChangeSetExecutionStatus == "UNAVAILABLE" || release.StackStatus == "ROLLBACK_COMPLETE") {
		// ChangeSet must be failed or Stack Status must be rolled back
		return nil
	}

	// DANGEROUS: should only get here if lots of checks pass
	return cf.DeleteStack(cfc, release.StackName)
}
