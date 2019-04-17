package template

import (
	"testing"

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
