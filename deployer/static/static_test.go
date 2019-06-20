package static

import (
	"encoding/base64"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/coinbase/fenrir/aws/mocks"
	"github.com/stretchr/testify/assert"
)

// To create this file, create the same directory structure
// Make sure to create these files using echo -n to prevent saving a newline:
//  `echo -n 'I AM ROOT FILE' > root`
//  `echo -n 'I AM FOLDER FILE' > folder/file`
//  `echo -n 'I AM NOT SVG' > folder/test.svg`
//  `echo -n '<svg></svg>' > folder/test2.svg`
//  `echo -n 'p { backround: black; }' > folder/test.css`
// To zip it:
//  `zip new.zip root folder/file folder/test.svg folder/test2.svg folder/test.css`
// To base64 encode it:
//  `base64 new.zip`
var SimpleZIPFile, _ = base64.StdEncoding.DecodeString(
	"UEsDBAoAAAAAAC+JSE5YS6enDgAAAA4AAAAEABwAcm9vdFVUCQADyrddXCi4XVx1eAsAAQT1AQAABBQAAABJIEFNIFJPT1QgRklMRVBLAwQKAAAAAADTiEhOXtqaYBAAAAAQAAAACwAcAGZvbGRlci9maWxlVVQJAAMdt11cH7ddXHV4CwABBPUBAAAEFAAAAEkgQU0gRk9MREVSIEZJTEVQSwMECgAAAAAA2ohITuHgaBoMAAAADAAAAA8AHABmb2xkZXIvdGVzdC5zdmdVVAkAAyu3XVwtt11cdXgLAAEE9QEAAAQUAAAASSBBTSBOT1QgU1ZHUEsDBBQAAAAIAOKISE4GH0AHCgAAAAsAAAAQABwAZm9sZGVyL3Rlc3QyLnN2Z1VUCQADOLddXCi4XVx1eAsAAQT1AQAABBQAAACzKS5Lt7PRB5EAUEsDBAoAAAAAAFeJSE4amjZ3FwAAABcAAAAPABwAZm9sZGVyL3Rlc3QuY3NzVVQJAAMVuF1cFbhdXHV4CwABBPUBAAAEFAAAAHAgeyBiYWNrcm91bmQ6IGJsYWNrOyB9UEsBAh4DCgAAAAAAL4lITlhLp6cOAAAADgAAAAQAGAAAAAAAAQAAAKSBAAAAAHJvb3RVVAUAA8q3XVx1eAsAAQT1AQAABBQAAABQSwECHgMKAAAAAADTiEhOXtqaYBAAAAAQAAAACwAYAAAAAAABAAAApIFMAAAAZm9sZGVyL2ZpbGVVVAUAAx23XVx1eAsAAQT1AQAABBQAAABQSwECHgMKAAAAAADaiEhO4eBoGgwAAAAMAAAADwAYAAAAAAABAAAApIGhAAAAZm9sZGVyL3Rlc3Quc3ZnVVQFAAMrt11cdXgLAAEE9QEAAAQUAAAAUEsBAh4DFAAAAAgA4ohITgYfQAcKAAAACwAAABAAGAAAAAAAAQAAAKSB9gAAAGZvbGRlci90ZXN0Mi5zdmdVVAUAAzi3XVx1eAsAAQT1AQAABBQAAABQSwECHgMKAAAAAABXiUhOGpo2dxcAAAAXAAAADwAYAAAAAAABAAAApIFKAQAAZm9sZGVyL3Rlc3QuY3NzVVQFAAMVuF1cdXgLAAEE9QEAAAQUAAAAUEsFBgAAAAAFAAUAmwEAAKoBAAAAAA==",
)

func Test_Custom_S3File(t *testing.T) {
	awsc := mocks.MockAWS()
	awsc.S3Client.AddGetObject("fromKey", "", nil)

	fn := staticSiteResources(awsc)
	id, data, err := fn(nil, cfn.Event{
		ResourceType: "Custom::S3File",
		RequestType:  "Create",
		StackID:      "arn:aws:cloudformation:us-east-1:00000000000:stack/stack/id",
		ResponseURL:  "http://localhost:8080",
		ResourceProperties: map[string]interface{}{
			"Bucket": "toBucket",
			"Key":    "toKey",
			"Uri":    "s3://fromBucket/fromKey",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, id, "toBucket/toKey")
	assert.Equal(t, data["Uri"], "s3://toBucket/toKey")
}

func Test_Custom_S3ZipFile(t *testing.T) {
	awsc := mocks.MockAWS()
	awsc.S3Client.AddGetObject("fromKey", string(SimpleZIPFile), nil)

	fn := staticSiteResources(awsc)
	id, data, err := fn(nil, cfn.Event{
		ResourceType: "Custom::S3ZipFile",
		RequestType:  "Create",
		StackID:      "arn:aws:cloudformation:us-east-1:00000000000:stack/stack/id",
		ResponseURL:  "http://localhost:8080",
		ResourceProperties: map[string]interface{}{
			"Bucket": "toBucket",
			"Key":    "toKey",
			"Uri":    "s3://fromBucket/fromKey",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, id, "toBucket/toKey")
	assert.Equal(t, len(data["Uris"].([]string)), 5)
}
