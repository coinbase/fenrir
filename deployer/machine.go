package deployer

import (
	"github.com/coinbase/fenrir/aws"
	"github.com/coinbase/fenrir/deployer/static"
	"github.com/coinbase/step/handler"
	"github.com/coinbase/step/machine"
)

// StateMachine returns the StateMachine for the deployer
func StateMachine() (*machine.StateMachine, error) {
	return machine.FromJSON([]byte(`{
    "Comment": "Fenrir Deployer",
    "StartAt": "Validate",
    "States": {
      "Validate": {
        "Type": "TaskFn",
        "Comment": "Validate and Set Defaults",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Lock",
        "Catch": [
          {
            "Comment": "Bad Release or Error GoTo end",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          }
        ]
      },
      "Lock": {
        "Type": "TaskFn",
        "Comment": "Grab Lock",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "CreateChangeSet",
        "Catch": [
          {
            "Comment": "Something else is deploying",
            "ErrorEquals": ["LockExistsError"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          },
          {
            "Comment": "Try Release Lock Then Fail",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "ReleaseLock"
          }
        ]
      },
      "CreateChangeSet": {
        "Type": "TaskFn",
        "Comment": "Create the CloudFormation ChangeSet",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "WaitForChangeSet",
        "Catch": [
          {
            "Comment": "Unsure of State, Fail",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "ReleaseLock"
          }
        ]
      },
      "WaitForChangeSet": {
        "Type": "Wait",
        "Seconds" : 5,
        "Next": "UpdateChangeSet"
      },
      "UpdateChangeSet": {
        "Type": "TaskFn",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Execute?",
        "Retry": [{
          "Comment": "Retry a few times in case of another error",
          "ErrorEquals": ["States.ALL"],
          "MaxAttempts": 3,
          "IntervalSeconds": 5
        }],
        "Catch": [{
          "Comment": "Cannot Handle this failure",
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "ReleaseLock"
        }]
      },
      "Execute?": {
        "Comment": "Wait until we are able to Execute",
        "Type": "Choice",
        "Choices": [
          {
            "Comment": "Continue to Success",
            "Variable": "$.change_set_execution_status",
            "StringEquals": "AVAILABLE",
            "Next": "Execute"
          },
          {
            "Comment": "It failed",
            "Variable": "$.change_set_status",
            "StringEquals": "FAILED",
            "Next": "ReleaseLock"
          }
        ],
        "Default": "WaitForChangeSet"
      },
      "Execute": {
        "Type": "TaskFn",
        "Comment": "Execute the Changeset",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "WaitForComplete",
        "Catch": [
          {
            "Comment": "Unsure of State, Fail",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "FailureDirty"
          }
        ]
      },
      "WaitForComplete": {
        "Type": "Wait",
        "Seconds" : 5,
        "Next": "UpdateStack"
      },
      "UpdateStack": {
        "Type": "TaskFn",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Complete?",
        "Retry": [{
          "Comment": "Retry a few times in case of another error",
          "ErrorEquals": ["States.ALL"],
          "MaxAttempts": 3,
          "IntervalSeconds": 5
        }],
        "Catch": [{
          "Comment": "Fail",
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "FailureDirty"
        }]
      },
      "Complete?": {
        "Comment": "End when $.stack_status enters end state",
        "Type": "Choice",
        "Choices": [
          {
            "OR": [
              { "Variable": "$.stack_status", "StringEquals": "CREATE_COMPLETE" },
              { "Variable": "$.stack_status", "StringEquals": "CREATE_FAILED" },
              { "Variable": "$.stack_status", "StringEquals": "DELETE_COMPLETE" },
              { "Variable": "$.stack_status", "StringEquals": "DELETE_FAILED" },
              { "Variable": "$.stack_status", "StringEquals": "ROLLBACK_COMPLETE" },
              { "Variable": "$.stack_status", "StringEquals": "ROLLBACK_FAILED" },
              { "Variable": "$.stack_status", "StringEquals": "UPDATE_COMPLETE" },
              { "Variable": "$.stack_status", "StringEquals": "UPDATE_ROLLBACK_COMPLETE" },
              { "Variable": "$.stack_status", "StringEquals": "UPDATE_ROLLBACK_FAILED" }
            ],
            "Next": "ReleaseLock"
          }
        ],
        "Default": "WaitForComplete"
      },
      "ReleaseLock": {
        "Type": "TaskFn",
        "Comment": "Release the Lock",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Success?",
        "Retry": [ {
          "Comment": "Keep trying to Release",
          "ErrorEquals": ["States.ALL"],
          "MaxAttempts": 3,
          "IntervalSeconds": 30
        }],
        "Catch": [{
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "FailureDirty"
        }]
      },
      "Success?": {
        "Comment": "Check the ChangeSet Complete",
        "Type": "Choice",
        "Choices": [
          {
            "OR": [
              { "Variable": "$.stack_status", "StringEquals": "CREATE_COMPLETE" },
              { "Variable": "$.stack_status", "StringEquals": "UPDATE_COMPLETE" }
            ],
            "Next": "Success"
          },
          {
            "OR": [
              { "Variable": "$.stack_status", "StringEquals": "ROLLBACK_FAILED" },
              { "Variable": "$.stack_status", "StringEquals": "UPDATE_ROLLBACK_FAILED" }
            ],
            "Next": "FailureDirty"
          }
        ],
        "Default": "CleanUp"
      },
      "CleanUp": {
        "Type": "TaskFn",
        "Comment": "Check if we need to clean anything up",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "FailureClean",
        "Retry": [{
          "Comment": "Keep trying to Release",
          "ErrorEquals": ["States.ALL"],
          "MaxAttempts": 3,
          "IntervalSeconds": 30
        }],
        "Catch": [{
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "FailureDirty"
        }]
      },
      "FailureClean": {
        "Comment": "Deploy Failed Cleanly",
        "Type": "Fail",
        "Error": "NotifyError"
      },
      "FailureDirty": {
        "Comment": "Deploy Failed, Resources left in Bad State, ALERT!",
        "Type": "Fail",
        "Error": "AlertError"
      },
      "Success": {
        "Type": "Succeed"
      }
    }
  }`))
}

// TaskHandlers returns
func TaskHandlers() *handler.TaskHandlers {
	return CreateTaskHandlers(&aws.ClientsStr{})
}

// CreateTaskHandlers returns
func CreateTaskHandlers(awsc aws.Clients) *handler.TaskHandlers {
	tm := handler.TaskHandlers{}

	tm[""] = static.StaticSiteResources(awsc)

	tm["Validate"] = Validate(awsc)
	tm["Lock"] = Lock(awsc)

	tm["CreateChangeSet"] = CreateChangeSet(awsc)
	tm["UpdateChangeSet"] = UpdateChangeSet(awsc)

	tm["Execute"] = Execute(awsc)

	tm["UpdateStack"] = UpdateStack(awsc)

	tm["ReleaseLock"] = ReleaseLock(awsc)

	tm["CleanUp"] = CleanUp(awsc)
	return &tm
}
