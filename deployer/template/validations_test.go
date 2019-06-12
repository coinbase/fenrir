package template

import (
	"testing"

	"github.com/awslabs/goformation/cloudformation/resources"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeName(t *testing.T) {
	norm := normalizeName("prefix", "project", "config", "resource", 64)

	assert.Equal(t, norm, "prefix-project-config-resource")

	norm = normalizeName("prefix", "project", "config", "long-resource-name-over-64-characters-limit", 64)
	norm2 := normalizeName("prefix", "project", "config", "long-resource-name-over-64-characters-limit-2", 64)

	assert.NotEqual(t, norm, norm2)
	assert.Equal(t, len(norm), 64)
	assert.Equal(t, len(norm2), 64)

	norm = normalizeName("prefix", "project", "config", "long-resource-name-over-64-characters-limit", 48)
	assert.Equal(t, len(norm), 48)
}

func TestValidateAWSServerlessSimpleTableWorks(t *testing.T) {
	template, err := MockTemplate("../../examples/tests/allowed/function.yml")
	assert.NoError(t, err)

	err = ValidateAWSServerlessSimpleTable("pn", "cn", "rn", template, &resources.AWSServerlessSimpleTable{})
	assert.NoError(t, err)
}

func TestValidateAWSServerlessLayerVersionWorks(t *testing.T) {
	template, err := MockTemplate("../../examples/tests/allowed/function.yml")
	assert.NoError(t, err)

	err = ValidateAWSServerlessLayerVersion("pn", "cn", "rn", template, &resources.AWSServerlessLayerVersion{
		ContentUri: "s3://bucket/path.zip",
	}, map[string]string{
		"s3://bucket/path.zip": MockS3SHA(),
	})
	assert.NoError(t, err)

}

func TestValidateAWSServerlessApiWorks(t *testing.T) {
	template, err := MockTemplate("../../examples/tests/allowed/function.yml")
	assert.NoError(t, err)

	err = ValidateAWSServerlessApi("pn", "cn", "rn", template, &resources.AWSServerlessApi{}, map[string]string{})
	assert.NoError(t, err)
}
