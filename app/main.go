package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

var Query = `
SELECT line_item_product_code,
	sum(line_item_blended_cost) AS cost,
	month
FROM %s
WHERE year='%v' AND month='%v'
GROUP BY  line_item_product_code, month
HAVING sum(line_item_blended_cost) > 0
ORDER BY  line_item_product_code;
`

var (
	S3_Bucket_Cost_And_Usage_RawData = os.Getenv("S3BucketCostAndUsageRawData")
	//
	SNS_Topic_Arn = os.Getenv("SNSTopicArn")
	//
	Athena_Database  = os.Getenv("AthenaDatabase")
	Athena_Workgroup = os.Getenv("AthenaWorkgroup")
	Athena_Table     = os.Getenv("AthenaTable")
	//
	Athena_Query_Result_Location = fmt.Sprintf("s3://%s/athena-output", S3_Bucket_Cost_And_Usage_RawData)
)

func handler(ctx context.Context) error {
	fmt.Println("S3_Bucket_Cost_And_Usage_RawData 	=>", S3_Bucket_Cost_And_Usage_RawData)
	fmt.Println("SNS_Topic_Arn 						=>", SNS_Topic_Arn)
	fmt.Println("Athena_Database 					=>", Athena_Database)
	fmt.Println("Athena_Workgroup 					=>", Athena_Workgroup)
	fmt.Println("Athena_Table 						=>", Athena_Table)
	fmt.Println("Athena_Query_Result_Location 		=>", Athena_Query_Result_Location)
	err := check_env()
	if err != nil {
		panic(err)
	}

	//
	year, month, _ := time.Now().UTC().Date()
	Query = fmt.Sprintf(Query, Athena_Table, year, int(month))
	fmt.Println("Query =>", Query)
	//
	queryExecutionId := athena_startQueryExecution(Query)
	fmt.Println("queryExecutionId =>", queryExecutionId)
	bucket, key := athena_getQueryExecution(queryExecutionId)
	fmt.Println("bucket, key =>", bucket, key)
	url, content := s3_getObjectAndPresignedURL(bucket, key)
	fmt.Println("url, content =>", url, content)
	sns_publish(url, content)

	return nil
}

func main() {
	lambda.Start(handler)
}

//
func check_env() (err error) {
	var env_issue = ""
	if len(S3_Bucket_Cost_And_Usage_RawData) == 0 {
		env_issue = "S3_Bucket_Cost_And_Usage_RawData"
	} else if len(SNS_Topic_Arn) == 0 {
		env_issue = "SNS_Topic_Arn"
	} else if len(Athena_Database) == 0 {
		env_issue = "Athena_Database"
	} else if len(Athena_Workgroup) == 0 {
		env_issue = "Athena_Workgroup"
	}
	if len(env_issue) != 0 {
		err = errors.New(fmt.Sprintf("%s is empty", env_issue))
	}
	return err
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
	resultConfiguration.SetOutputLocation(Athena_Query_Result_Location)
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
	//
START:
	time.Sleep(time.Second * 2)
	//
	input := athena.GetQueryExecutionInput{}
	input.SetQueryExecutionId(queryExecutionId)
	output, err := svc_athena.GetQueryExecution(&input)
	if err != nil {
		panic(err)
	}

	text := fmt.Sprintf("athena query State => %s\n", *output.QueryExecution.Status.State)
	if strings.ToLower(*output.QueryExecution.Status.State) == "failed" {
		panic(errors.New(text))
	} else if strings.ToLower(*output.QueryExecution.Status.State) != "succeeded" {
		fmt.Println(text)
		goto START
	}

	tmp := strings.ReplaceAll(*output.QueryExecution.ResultConfiguration.OutputLocation, "s3://", "")
	s3_location := strings.SplitN(tmp, "/", 2)
	s3_bucket = s3_location[0]
	s3_key = s3_location[1]
	return
}

// s3
var svc_s3 = s3.New(mySession)

func s3_getObjectAndPresignedURL(bucket, key string) (urlStr, contentStr string) {
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

func sns_publish(url, content string) {
	//
	msg_content := fmt.Sprintf("```\n%s\n```", strings.TrimSpace(content))
	msg_url := fmt.Sprintf("<%s|download_link>", url)
	final_msg := fmt.Sprintf("*Hello, this is Cost And Usage Report*\n\n%s\n%s", msg_content, msg_url)
	//
	input := sns.PublishInput{}
	input.SetMessage(final_msg)
	input.SetSubject("Cost And Usage Report")
	input.SetTopicArn(SNS_Topic_Arn)

	output, err := svc_sns.Publish(&input)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
