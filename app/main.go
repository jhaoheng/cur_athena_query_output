package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

var (
	SNS_Topic_Arn                   = os.Getenv("SNS_Topic_Arn")
	S3Bucket_Of_AthenaReport_Output = os.Getenv("S3Bucket_Of_AthenaReport_Output")
	Athena_Database                 = os.Getenv("Athena_Database")
	Athena_Workgroup                = os.Getenv("Athena_Workgroup")
	Querys                          = []string{
		`SELECT line_item_product_code,
					sum(line_item_blended_cost) AS cost,
					month
			FROM all
			WHERE year='2020'
				AND month='11'
			GROUP BY  line_item_product_code, month
			HAVING sum(line_item_blended_cost) > 0
			ORDER BY  line_item_product_code;`,
	}
)

func handler(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		s3 := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3.Bucket.Name, s3.Object.Key)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}

// athena
var mySession = session.Must(session.NewSession())
var svc_athena = athena.New(mySession)

func athena_startQueryExecution(query string) (queryExecutionId string) {
	//
	queryExecutionContext := athena.QueryExecutionContext{}
	queryExecutionContext.SetDatabase(Athena_Database)
	//
	resultConfiguration := athena.ResultConfiguration{}
	resultConfiguration.SetOutputLocation(S3Bucket_Of_AthenaReport_Output)
	//
	input := athena.StartQueryExecutionInput{
		QueryString:           aws.String(query),
		WorkGroup:             aws.String(Athena_Workgroup),
		QueryExecutionContext: &queryExecutionContext,
		ResultConfiguration:   &resultConfiguration,
	}
	output, err := svc_athena.StartQueryExecution(&input)
	if err != nil {
		panic(err)
	}
	queryExecutionId = *output.QueryExecutionId
	return
}

func athena_getQueryExecution(queryExecutionId string) (s3_bucket, s3_key string) {
	input := athena.GetQueryExecutionInput{}
	input.SetQueryExecutionId(queryExecutionId)
	output, err := svc_athena.GetQueryExecution(&input)
	if err != nil {
		panic(err)
	}

	if *output.QueryExecution.Status.State != "SUCCEEDED" {
		err := errors.New("athena query fail")
		panic(err)
	}
	tmp := strings.ReplaceAll(*output.QueryExecution.ResultConfiguration.OutputLocation, "s3://", "")
	s3_location := strings.SplitN(tmp, "/", 2)
	s3_bucket = s3_location[0]
	s3_key = s3_location[1]
	return
}

// s3
var svc_s3 = s3.New(mySession)

func s3_GetObjectAndPresignedURL(bucket, key string) (urlStr, contentStr string) {
	req, result := svc_s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	err := req.Send()
	if err != nil {
		panic(err)
	}
	//
	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	contentStr = buf.String()
	//
	urlStr, err = req.Presign(24 * time.Hour)
	if err != nil {
		panic(err)
	}
	return urlStr, contentStr
}

//
var svc_sns = sns.New(mySession)

func sns_publish(message string) {
	input := sns.PublishInput{}
	input.SetMessage(message)
	input.SetSubject("Cost And Usage Report")
	input.SetTopicArn(SNS_Topic_Arn)

	output, err := svc_sns.Publish(&input)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
