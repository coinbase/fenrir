package mocks

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/step/utils/to"
)

// CFClient returns
type CFClient struct {
	aws.CFAPI
}

func (m *CFClient) init() {}

// DescribeStacks returns
func (m *CFClient) DescribeStacks(in *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	return &cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{
			&cloudformation.Stack{
				StackStatus: to.Strp("CREATE_COMPLETE"),
			},
		},
	}, nil
}

// DescribeStacks returns
func (m *CFClient) CreateChangeSet(in *cloudformation.CreateChangeSetInput) (*cloudformation.CreateChangeSetOutput, error) {
	return nil, nil
}

// ExecuteChangeSet returns
func (m *CFClient) ExecuteChangeSet(in *cloudformation.ExecuteChangeSetInput) (*cloudformation.ExecuteChangeSetOutput, error) {
	return nil, nil
}

// DescribeChangeSet returns
func (m *CFClient) DescribeChangeSet(in *cloudformation.DescribeChangeSetInput) (*cloudformation.DescribeChangeSetOutput, error) {
	return &cloudformation.DescribeChangeSetOutput{
		Status:          to.Strp("CREATE_COMPLETE"),
		ExecutionStatus: to.Strp("AVAILABLE"),
	}, nil
}

func (m *CFClient) DescribeStackEvents(in *cloudformation.DescribeStackEventsInput) (*cloudformation.DescribeStackEventsOutput, error) {
	return &cloudformation.DescribeStackEventsOutput{
		StackEvents: []*cloudformation.StackEvent{
			&cloudformation.StackEvent{
				Timestamp:            &time.Time{},
				ResourceStatus:       to.Strp("ResourceStatus"),
				ResourceType:         to.Strp("ResourceType"),
				LogicalResourceId:    to.Strp("LogicalResourceId"),
				ResourceStatusReason: to.Strp("ResourceStatusReason"),
			},
		},
	}, nil
}
