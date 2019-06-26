package custom

import (
  "fmt"

  "github.com/coinbase/step/aws"
  "github.com/coinbase/step/aws/s3"
)

func handleS3File(bucket string, key string, uri string, lambdas3c, s3c aws.S3API) (string, map[string]interface{}, error) {
  physicalResourceID := fmt.Sprintf("%v/%v", bucket, key)
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

  contentType := detectContentType(key, *obj)
  err = s3.PutWithType(s3c, &bucket, &key, obj, &contentType)
  if err != nil {
    return physicalResourceID, data, err
  }

  data["Uri"] = s3Uri(bucket, key)

  return physicalResourceID, data, nil
}
