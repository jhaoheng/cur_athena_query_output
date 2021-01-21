# 不使用 S3 Event Notification 自動執行 Athena Query

## 原因一 : SAM 無法對已經存在的 s3 bucket, 建立事件通知
> it has clearly said "to specify an S3 bucket as an event source for a Lambda function, both resources have to be declared in the same template. AWS SAM does not support specifying an existing bucket as an event source."
https://github.com/awslabs/serverless-application-model/blob/master/versions/2016-10-31.md#example-awsserverlessfunctio

- 所以建立好 lambda 後, 必須至 cost-and-usage 原始資料存放的 bucket 中, 建立 **event notification**

## 原因二 : 若使用 CloudWatch Event 來觸發 S3 事件
- 必須建立 Trail, 會有額外的支出 (預設只有一個免費)

## 原因三 : s3 Event Notification 無法存在兩個相同的事件觸發
- 事件觸發用於匯入資料到 Athena
