package static

import (
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/coinbase/fenrir/aws/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_staticSiteResources(t *testing.T) {
	awsc := mocks.MockAWS()
	fn := staticSiteResources(awsc)
	id, data, err := fn(nil, cfn.Event{
		RequestID:          "asd",
		LogicalResourceID:  "asd",
		StackID:            "asd",
		ResponseURL:        "http://localhost:8080",
		ResourceProperties: map[string]interface{}{"Echo": "asd"},
	})
	assert.NoError(t, err)
	assert.Equal(t, id, "asd")
	assert.Equal(t, data["Echo"], "asd")
}
