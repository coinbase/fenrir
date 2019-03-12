package deployer

import (
	"context"
	"fmt"

	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/bifrost"
	"github.com/coinbase/step/errors"
	"github.com/coinbase/step/utils/to"
)

// DeployHandler function type
type DeployHandler func(context.Context, *Release) (*Release, error)

////////////
// HANDLERS
////////////

var assumedRole = to.Strp("coinbase-fenrir-assumed")

// Validate checks the release
func Validate(awsc aws.Clients) DeployHandler {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// Assign the release its SHA before anything alters it
		release.ReleaseSHA256 = to.SHA256Struct(release)

		// Default the releases Account and Region to where the Lambda is running
		region, account := to.AwsRegionAccountFromContext(ctx)

		// Fill in all the blank Attributes
		release.SetDefaults(region, account)

		if err := release.Validate(awsc.S3(nil, nil, nil)); err != nil {
			return nil, &errors.BadReleaseError{err.Error()}
		}

		if err := release.ValidateTemplate(
			awsc.EC2(release.AwsRegion, release.AwsAccountID, assumedRole),
			awsc.IAM(release.AwsRegion, release.AwsAccountID, assumedRole),
			awsc.S3(release.AwsRegion, release.AwsAccountID, assumedRole),
			awsc.KIN(release.AwsRegion, release.AwsAccountID, assumedRole),
			awsc.DDB(release.AwsRegion, release.AwsAccountID, assumedRole),
			awsc.SQS(release.AwsRegion, release.AwsAccountID, assumedRole),
		); err != nil {
			return nil, &errors.BadReleaseError{err.Error()}
		}

		return release, nil
	}
}

// Lock secures a lock for the release
func Lock(awsc aws.Clients) interface{} {
	return func(ctx context.Context, release *Release) (*Release, error) {
		// returns LockExistsError, LockError
		return release, release.GrabLock(awsc.S3(nil, nil, nil))
	}
}

// CreateChangeSet crates new AWS resources for the release
func CreateChangeSet(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {

		if err := release.CreateChangeSet(
			awsc.CF(release.AwsRegion, release.AwsAccountID, assumedRole),
		); err != nil {
			return nil, &errors.BadReleaseError{err.Error()}
		}

		return release, nil
	}
}

func UpdateChangeSet(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {
		err := release.FetchChangeSet(awsc.CF(release.AwsRegion, release.AwsAccountID, assumedRole))

		if err != nil {
			return nil, err
		}

		return release, nil
	}
}

// Execute executes the changeset
func Execute(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {
		if err := release.Execute(
			awsc.CF(release.AwsRegion, release.AwsAccountID, assumedRole),
		); err != nil {
			return nil, &errors.HaltError{err.Error()}
		}

		return release, nil
	}
}

// ReleaseLock releases lock with sucess
func ReleaseLock(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {

		if err := release.ReleaseLock(awsc.S3(nil, nil, nil)); err != nil {
			return nil, &errors.LockError{err.Error()}
		}

		release.Success = to.Boolp(true)

		return release, nil
	}
}

func UpdateStack(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {
		err := release.FetchStack(
			awsc.S3(nil, nil, nil),
			awsc.CF(release.AwsRegion, release.AwsAccountID, assumedRole),
		)

		if err != nil {
			return nil, err
		}

		return release, nil
	}
}

// ReleaseLockFailure releases the lock then fails
func CleanUp(awsc aws.Clients) DeployHandler {
	return func(_ context.Context, release *Release) (*Release, error) {

		release.Success = to.Boolp(false)

		if err := release.CleanUp(
			awsc.S3(nil, nil, nil),
			awsc.CF(release.AwsRegion, release.AwsAccountID, assumedRole),
		); err != nil {
			return nil, &errors.CleanUpError{err.Error()}
		}

		// Add Error if if can be found
		if release.Error == nil {
			cause := ""
			if release.ChangeSetStatusReason != "" {
				cause += fmt.Sprintf("changeset: %s", release.ChangeSetStatusReason)
			}

			if release.StackStatusReason != "" && release.StackStatusReason != "User Initiated"{
				cause += fmt.Sprintf(":: stack: %s", release.StackStatusReason)
			}

			release.Error = &bifrost.ReleaseError{
				Error: to.Strp("Failed"),
				Cause: &cause,
			}
		}

		return release, nil
	}
}
