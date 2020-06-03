module github.com/coinbase/fenrir

require (
	github.com/aws/aws-lambda-go v1.17.0
	github.com/aws/aws-sdk-go v1.31.9
	github.com/awslabs/goformation/v4 v4.8.0
	github.com/coinbase/step v1.0.2
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/rogpeppe/godef v1.1.2 // indirect
	github.com/sanathkr/yaml v0.0.0-20170819201035-0056894fa522
	github.com/stretchr/testify v1.5.1
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/tools v0.0.0-20200601175630-2caf76543d99 // indirect
)

// This replaces goformation with a fork that has the fix on it
// TODO replace once PR https://github.com/awslabs/goformation/pull/271 is merged
replace github.com/awslabs/goformation/v4 v4.6.0 => github.com/grahamjenson/goformation/v4 v4.0.0-20200227205046-704c8e4046a8

go 1.13
