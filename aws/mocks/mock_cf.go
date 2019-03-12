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
	StackResp *cloudformation.DescribeStacksOutput
	ChangeSet *cloudformation.DescribeChangeSetOutput
	DeleteStackCalled bool
}

func (m *CFClient) init() {
	if m.StackResp == nil {
		m.StackResp = &cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus:  to.Strp("CREATE_COMPLETE"),
					CreationTime: to.Timep(time.Now()),
				},
			},
		}
	}

	if m.ChangeSet == nil {
		m.ChangeSet = &cloudformation.DescribeChangeSetOutput{
			Status:          to.Strp("CREATE_COMPLETE"),
			ExecutionStatus: to.Strp("AVAILABLE"),
		}
	}
}

// DescribeStacks returns
func (m *CFClient) DescribeStacks(in *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	m.init()
	return m.StackResp, nil
}

// DescribeStacks returns
func (m *CFClient) CreateChangeSet(in *cloudformation.CreateChangeSetInput) (*cloudformation.CreateChangeSetOutput, error) {
	return nil, nil
}

// ExecuteChangeSet returns
func (m *CFClient) ExecuteChangeSet(in *cloudformation.ExecuteChangeSetInput) (*cloudformation.ExecuteChangeSetOutput, error) {
	return nil, nil
}

// DeleteStack returns
func (m *CFClient) DeleteStack(in *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	m.DeleteStackCalled = true
	return nil, nil
}

// DescribeChangeSet returns
func (m *CFClient) DescribeChangeSet(in *cloudformation.DescribeChangeSetInput) (*cloudformation.DescribeChangeSetOutput, error) {
	m.init()
	return m.ChangeSet, nil
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
