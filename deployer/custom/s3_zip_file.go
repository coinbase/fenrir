package custom

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/coinbase/step/aws"
	"github.com/coinbase/step/aws/s3"
)

func handleS3ZipFile(bucket string, key string, uri string, lambdas3c, s3c aws.S3API) (string, map[string]interface{}, error) {
	physicalResourceID := fmt.Sprintf("%v/%v", bucket, key)
	uris := []string{}
	data := map[string]interface{}{}

	fromBucket, fromKey, err := s3UriToBucketKey(uri)

	if err != nil {
		return physicalResourceID, data, err
	}

	// Get Zip File
	obj, err := s3.Get(lambdas3c, &fromBucket, &fromKey)
	if err != nil {
		return physicalResourceID, data, err
	}

	zippedReader, err := zip.NewReader(bytes.NewReader(*obj), int64(len(*obj)))
	if err != nil {
		return physicalResourceID, data, err
	}

	for _, zf := range zippedReader.File {
		src, err := zf.Open()
		if err != nil {
			return physicalResourceID, data, err
		}
		defer src.Close()

		if zf.FileInfo().IsDir() {
			continue // S3 just needs path
		}

		read, err := ioutil.ReadAll(src)
		if err != nil {
			return physicalResourceID, data, err
		}

		contentType := detectContentType(zf.Name, read)

		key := fmt.Sprintf("%v%v", key, zf.Name)

		err = s3.PutWithType(s3c, &bucket, &key, &read, &contentType)
		if err != nil {
			return physicalResourceID, data, err
		}

		uris = append(uris, s3Uri(bucket, key))
	}
	data["Uris"] = uris
	return physicalResourceID, data, nil
}
