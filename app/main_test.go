package main

import (
	"fmt"
	"testing"
)

func Test_athena_startQueryExecution(t *testing.T) {
	Athena_Database = "athenacurcfn_all"
	Athena_Query_Result_Location = "s3://aws-athena-query-results-424613967558-ap-southeast-1/result/"
	Athena_Workgroup = "primary"

	for _, query := range Querys {
		queryExecutionId := athena_startQueryExecution(query)
		fmt.Println(queryExecutionId)
	}
}

func Test_athena_getQueryExecution(t *testing.T) {
	queryExecutionId := "19d79aac-f3ce-4255-8920-2b39d8f3178c"
	s3_bucket, s3_key := athena_getQueryExecution(queryExecutionId)
	fmt.Println(s3_bucket)
	fmt.Println(s3_key)
}

func Test_s3_getObjectAndPresignedURL(t *testing.T) {
	bucket := "aws-athena-query-results-424613967558-ap-southeast-1"
	key := "result/19d79aac-f3ce-4255-8920-2b39d8f3178c.csv"
	url, content := s3_getObjectAndPresignedURL(bucket, key)

	fmt.Println(url)
	fmt.Println(content)
}

func Test_sns_publish(t *testing.T) {
	var content, url string

	//
	url = "http://google.com"
	content = `
"line_item_product_code","cost","month"
"AWSGlue","0.12075812","11"
"AWSSystemsManager","5.070000000000001E-4","11"
"AmazonAthena","5.0E-4","11"
"AmazonEC2","0.49475327289999993","11"
"AmazonECR","0.0022787006999999996","11"
"AmazonRDS","0.035728597099999995","11"
"AmazonRoute53","0.530644","11"
"AmazonS3","0.03333862299999986","11"
"AmazonVPC","9.254024056399999","11"
	`

	//
	SNS_Topic_Arn = "arn:aws:sns:ap-southeast-1:424613967558:SNSToLambdaToSlack-INFO"
	//
	sns_publish(url, content)
}
