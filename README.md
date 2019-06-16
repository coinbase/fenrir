# Fenrir

<img src="./assets/logo.png" align="right" alt="Odin" />

Fenrir is a secure AWS SAM deployer that can help manage your own serverless projects or scale serverless to a large organization. At its core it is a reimplementation of the `sam deploy` command as an [AWS Step Function](https://blog.coinbase.com/aws-step-functions-state-machines-bifrost-and-building-deployers-5e3745fe645b?gi=fd665a0a4039), so it's a serverless serverless (*serverless^2*) deployer. Fenrir also:

1. **Uses consistent naming:** good naming (and tagging) of resources, like Lambda and API Gateway, will keep accounts clean and make obvious which resources belong to which projects.
1. **Follows recommended security practices:** e.g. practice "least privilege" by giving Lambdas separate security groups and IAM roles.
1. **Creates a reliable workflow:** cleanly handle failure in a way that shows what happened, why it happened, and how to remedy.
1. **Records what is deployed:** quickly answering what is currently deployed allows engineers to debug and understand the current state of the world.

The goal is to provide a secure and pleasant experience for building and deploying serverless applications that can be used by a single developer or a large organisation.

## Getting Started

Deploy Fenrir to AWS using `./scripts/cf_bootstrap <s3_bucket>`. This creates a CloudFormation stack with the Fenrir Step Function, Lambdas, Buckets and Roles fenrir needs to run.

You can then `cd examples/hello` and `fenrir package && fenrir deploy` which will deploy the `hello` example application.

### Hello application

Fenrir supports a subset of AWS SAM templates with only the addition of adding `ProjectName` and `ConfigName` to the top of the template.

The hello application `template.yml` looks like:

```
ProjectName: "coinbase/deploy-test"
ConfigName: "development"

AWSTemplateFormatVersion: "2010-09-09"
Transform: AWS::Serverless-2016-10-31

Resources:
  helloAPI:
    Type: AWS::Serverless::Api
    Properties:
      StageName: dev
      EndpointConfiguration: REGIONAL
  hello:
    Type: AWS::Serverless::Function
    Properties:
      CodeUri: .
      Role: lambda-role
      Handler: hello.lambda
      Runtime: go1.x
      Events:
        hi:
          Type: Api
          Properties:
            RestApiId: !Ref helloAPI
            Path: /hello
            Method: GET
```

With code that looks like:

```
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(func(_ interface{}) (interface{}, error) {
		return map[string]string{"body": "Hello"}, nil
	})
}
```

The name of the lambda function is `hello` so Fenrir expects the file `/hello.zip` to exist in the built docker conatiner by having a Dockerfile:

```
FROM golang
WORKDIR /
RUN apt-get update && apt-get upgrade -y && apt-get install -y zip

COPY . .
RUN go get github.com/aws/aws-lambda-go/lambda
RUN GOOS=linux GOARCH=amd64 go build -o hello.lambda .
RUN zip hello.zip hello.lambda
```

With these in place you can now execute:

* `go build -o hello.lambda . && sam local start-api` to start a local test API
* `fenrir package` to prepare the files needed to deploy
* `fenrir deploy` to deploy the template (*requires fenrir deployer*)

## Supported Resources

Fenrir does not support all SAM resources or all properties. Generally it limits all references resources (e.g. Security Groups, Subnets, S3, Kinesis) to have specific tags AND it forces good naming patterns to stop conflicts.

The specific resources that it supports, and their limitations are:

### AWS::Serverless::Function

1. `FunctionName` is generated and cannot be defined.
1. `VPCConfig.SecurityGroupIds` Each SG must have the `ProjectName`, `ConfigName` same as the template, and `ServiceName` equal to the name of the Lambda resource.
1. `VPCConfig.SubnetIds` must have the `DeployWithFenrir` tag equal to `true`.
1. `Role` must have the tags `ProjectName`, `ConfigName` same as the template, and `ServiceName` equal to the name of the Lambda resource.
1. `PermissionsBoundary` must be defined, is defaulted to `fenrir-permissions-boundary`, must have *correct tags* (**TODO** for now it is hard coded as default)
1. `Policies` only supports a list of SAM Policy templates of type (w/ limitations):
	1. `DynamoDBCrudPolicy` where `TableName` must be a local `!Ref`
	1. `LambdaInvokePolicy` where `FunctionName` must be a local `!Ref`
	1. `KMSDecryptPolicy` where ref'd `KeyId` (can be alias) must have *correct tags*
	1. `VPCAccessPolicy` 
1. `Events` supported `Type`s and their limitations are:
	1. `Api`: It must have `RestApiId` that is a reference to a local API resource
	1. `S3`: `Bucket` must have *correct tags*<sup>*</sup>
	1. `Kinesis`: `Stream` must have *correct tags*<sup>*</sup>
	1. `DynamoDB`: `Stream` must have *correct tags*<sup>*</sup>
	1. `SQS`: `Queue` must have *correct tags*<sup>*</sup>
 	1. `SNS`: `Topic` can be topic name or ARN and must have *correct tags*<sup>*</sup>
	1. `Schedule`
	1. `CloudWatchEvent`

<sup>*</sup>: *correct tags* means tags are `FenrirAllAllowed=true` OR have `FenrirAllowed:<project>:<config>=true` OR `ProjectName` and `ConfigName` tags equal to the release.

### AWS::Serverless::Api

The limitations are:

1. `Name` is generated and cannot be defined
1. `EndpointConfiguration` defaults to `PRIVATE`

### AWS::Serverless::LayerVersion

The limitations are:

1. `LayerName` is generated and cannot be defined

### AWS::Serverless::SimpleTable

1. `TableName` is generated and cannot be defined
2. `DeletionPolicy` is defaulted to `Retain`


## Fenrir Deployer

Fenrir is a [Bifrost Step Function](https://github.com/coinbase/bifrost) reimplemetnation of `aws cloudformation deploy` [script](https://github.com/aws/aws-cli/blob/master/awscli/customizations/cloudformation/deployer.py). The logic flow looks like:

<img src="./assets/sm.png" alt="state diagram"/>

## TODOs

There is always more to do:

1. Auto add common sense Outputs
1. S3 Static site uploader
1. Support Role Arns and Name Tags
1. Layers should not include environment e.g. development, just configuration to be the same ARN across accounts
1. Layers should be able to reference "latest" version
1. Let Fenrir Bootstrap itself by letting it deploy Step Functions

## More Links

Links I have found useful:

https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-private-apis.html

API gateway resource policy:
https://github.com/awslabs/serverless-application-model/issues/514
