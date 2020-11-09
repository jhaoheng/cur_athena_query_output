## 目的
- 建立完成 cost and usage report 後
- 透過事件的觸發, 讓 lambda 執行既定的 query, 產出報告後, 送到 sns 中
- 可透過其他方式從 sns 中訂閱, 取得報告

## 建立順序
1. 建立 cost-and-usage report
    - [ ] : 勾選 `Include resource IDs`
    - [ ] : 建立資料存放的 bucket
    - [ ] : *Enable report data integration for* 設定 Athena
        - 建立 template : `https://docs.aws.amazon.com/cur/latest/userguide/cur-ate-setup.html`
2. 建立 cost-and-usage 的 cfn template : `https://docs.aws.amazon.com/cur/latest/userguide/cur-ate-setup.html`
    1. 需等待一段時間, 約八個小時, 等待 AWS 產生資料
    2. 到 bucket 中, 可找到 .yml 的 cfn 設定檔
    3. 下載後建立 cfn
3. 取得以下參數, 並建立此 github 服務, 目的 : 執行指定 query, 並輸出到 sns
    - Athena
        - AthenaDatabase
        - AthenaWorkgroup
    - S3BucketCostAndUsageRawData
    - SNSTopicArn
4. 設定 s3 中建立 
    - event notification
        1. properties -> Event notifications
        2. 設定 sufix = .csv
        3. 設定 event type = Put
        4. destination = lambda
    - 設定 Lifecycle
        1. Management -> Lifecycle rules
5. 測試
    - 上傳 .csv 到 bucket 中

## 本地測試
1. 設定環境
    - 確定 cloud s3 / athena / cost-and-usage report 均設定完成
    - `export AWS_REGION=ap-southeast-1`
    - `export AWS_PROFILE=default`
2. 執行 lambda


# 注意 : 關於自動化

## SAM 無法對已經存在的 s3 bucket, 建立事件通知

> it has clearly said "to specify an S3 bucket as an event source for a Lambda function, both resources have to be declared in the same template. AWS SAM does not support specifying an existing bucket as an event source."
https://github.com/awslabs/serverless-application-model/blob/master/versions/2016-10-31.md#example-awsserverlessfunctio

- 所以建立好 lambda 後, 必須至 cost-and-usage 原始資料存放的 bucket 中, 建立 **event notification**

## 使用 CloudWatch Event 來觸發 S3 事件
- 必須建立 Trail, 會有額外的支出 (預設只有一個免費)

 