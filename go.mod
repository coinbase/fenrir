module github.com/coinbase/fenrir

require (
	github.com/aws/aws-lambda-go v1.11.1
	github.com/aws/aws-sdk-go v1.20.2
	github.com/awslabs/goformation v0.0.0-20190611064747-e853a3c93929
	github.com/coinbase/odin v0.0.0-20190410082300-c7799f955f38
	github.com/coinbase/step v0.0.0-20190408131218-9f799639d07c
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/sanathkr/yaml v0.0.0-20170819201035-0056894fa522
	github.com/stretchr/testify v1.3.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.1.0
)

replace github.com/awslabs/goformation => ../../awslabs/goformation

replace github.com/coinbase/step => ../step
