package main

import (
	"fmt"
	"testing"
)

func Test_athena_getQueryExecution(t *testing.T) {
	queryExecutionId := "19d79aac-f3ce-4255-8920-2b39d8f3178c"
	s3_bucket, s3_key := athena_getQueryExecution(queryExecutionId)
	fmt.Println(s3_bucket)
	fmt.Println(s3_key)
}

func Test_s3_GetObjectAndPresignedURL(t *testing.T) {
	bucket := "aws-athena-query-results-424613967558-ap-southeast-1"
	key := "result/19d79aac-f3ce-4255-8920-2b39d8f3178c.csv"
	url, content := s3_GetObjectAndPresignedURL(bucket, key)

	fmt.Println(url)
	fmt.Println(content)
}
