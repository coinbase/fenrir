package client

import (
	"testing"

	"github.com/coinbase/step/utils/to"
	"github.com/stretchr/testify/assert"
)

func Test_Successful_Client_Parsing(t *testing.T) {
	release, err := releaseFromFile(
		to.Strp("../examples/tests/allowed/hello.yml"),
		to.Strp("region"),
		to.Strp("000"),
	)

	assert.NoError(t, err)
	assert.Equal(t, *release.AwsAccountID, "000")
	assert.Equal(t, *release.AwsRegion, "region")
}
