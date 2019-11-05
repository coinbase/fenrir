module github.com/coinbase/fenrir

require (
	github.com/aws/aws-lambda-go v1.11.1
	github.com/aws/aws-sdk-go v1.20.2
	github.com/awslabs/goformation/v3 v3.0.0
	github.com/coinbase/step v0.0.0-20190913195516-3d78393342f3
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/sanathkr/yaml v0.0.0-20170819201035-0056894fa522
	github.com/stretchr/testify v1.4.0
	github.com/xeipuuv/gojsonschema v1.1.0
)

# TODO replace once PR is merged
replace github.com/awslabs/goformation/v3 v3.0.0 => github.com/grahamjenson/goformation/v3 v3.0.0-20191105231909-547d63e1fd68

go 1.13
