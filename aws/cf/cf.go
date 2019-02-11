package cf

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

type CreateChangeSetInput cloudformation.CreateChangeSetInput

type NotFoundError struct {
	s string
}

func (e NotFoundError) Error() string {
	return e.s
}

func DescribeStack(cfc aws.CFAPI, name *string) (*cloudformation.Stack, error) {
	output, err := cfc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: name,
	})

	if err != nil {
		return nil, err
	}

	if output == nil || output.Stacks == nil {
		return nil, fmt.Errorf("DescribeStack: Unknown CF error")
	}

	if len(output.Stacks) != 1 {
		return nil, NotFoundError{"DescribeStack: Not Found"}
	}

	if output.Stacks[0] == nil {
		return nil, fmt.Errorf("DescribeStack: Unknown CF error")
	}

	return output.Stacks[0], nil
}

func DeleteStack(cfc aws.CFAPI, name *string) error {
	_, err := cfc.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: name,
	})

	return err
}

// HasStack Hack from https://github.com/aws/aws-cli/blob/master/awscli/customizations/cloudformation/deployer.py#L38
func HasStack(cfc aws.CFAPI, name *string) (bool, error) {
	output, err := cfc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: name,
	})

	if err != nil {
		// Hackity explained in above link
		if strings.Contains(err.Error(), fmt.Sprintf("Stack with id %s does not exist", *name)) {
			return false, nil
		}
		return false, err
	}

	if output == nil || output.Stacks == nil {
		return false, fmt.Errorf("HasStack: Unknown CF error")
	}

	if len(output.Stacks) != 1 {
		return false, nil
	}

	if output.Stacks[0] == nil || output.Stacks[0].StackStatus == nil {
		return false, fmt.Errorf("HasStack: Unknown CF error")
	}

	// Hackity explained in above link
	return *output.Stacks[0].StackStatus != "REVIEW_IN_PROGRESS", nil
}

func ChangeSetType(cfc aws.CFAPI, stackName *string) (*string, error) {
	hasStack, err := HasStack(cfc, stackName)
	if err != nil {
		return nil, err
	}

	if hasStack {
		return to.Strp("UPDATE"), nil
	}

	return to.Strp("CREATE"), nil
}

func CreateChangeSet(cfc aws.CFAPI, input *cloudformation.CreateChangeSetInput) error {
	// Try Create
	_, err := cfc.CreateChangeSet(input)
	return err
}
