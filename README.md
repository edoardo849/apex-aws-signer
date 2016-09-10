![go-version: 1.7](https://img.shields.io/badge/go--version-1.7-blue.svg)
# AWS Signer for Apex
A roundTripper implementation for Go to sign http requests to AWS services from AWS Lambda.

## Requirements
This library is supposed to be used on [AWS Lambda](https://aws.amazon.com/lambda/) with the [go-apex](https://github.com/apex/go-apex) library / framework.

## Background
The AWS SDK is a great tool to perform authenticated requests to AWS services. However, when trying to access safely the applications hosted on those services, it becomes necessary to sign the request. A great upside of this method is that signed requests allows a resource to be authenticated by its IAM role. You can therefore assign and revoke IAM policies to the role  keeping your infrastructure safe from the inside and the outside.

### Use case: AWS Lambda to ElasticSearch with Go
This library was first built to access [AWS ElasticSearch](https://aws.amazon.com/elasticsearch-service/) from a Lambda function. The use case is the following: I have an S3 bucket that triggers a Lambda function when a new item is created or updated. The Lambda function pulls the object from the bucket performing some transformation and then ingests its content into an ElasticSearch index. This last part can be performed **safely** in two ways:

1. You can create a VPC with your Lambda function and the ElasticSearch service and restrict access to ES to the privateIP of the VPC 
1. You can still create a VPC for security purposes but you can restrict even more the ES access to just an IAM role. Apex will automatically create that role for you, so you can attach to it a policy to access ES.

The first example can be good enough in most scenarios. However, one important downsize is that your ES instance will be still open to every resource inside the VPC, which is not great because a good security strategy protects both against external and internal threats.

# Usage

Install the library as usual:

```bash 
go get github.com/edoardo849/apex-aws-signer
```

If you want to run the tests, just in case... :

```bash
cd $GOPATH/src/github.com/edoardo849/apex-aws-signer
go test -cover
```

For example, if you're using ElasticSearch with @olivere 's elastic library:

```go 
import (
    "github.com/edoardo849/apex-aws-signer"
    "github.com/apex/log"
    "github.com/aws/aws-sdk-go/service/elasticsearchservice"
    "gopkg.in/olivere/elastic.v3"
)

// Example For ElasticSearch
// ctx is the *apex.Context
ctxLogger := log.WithField("requestID", ctx.RequestID)
transport := NewAWSSigningTransport(s, elasticsearchservice.ServiceName)
transport.Logger = ctxLogger

httpClient := &http.Client{
	Transport: transport,
}
// Use the client with Olivere's elastic client
client, err := elastic.NewClient(
    elastic.SetSniff(false),
    elastic.SetHealthcheckTimeout(time.Duration(2)*time.Second),
    elastic.SetURL("your-aws-es-endpoint"),
    elastic.SetScheme("https"),
    elastic.SetHttpClient(httpClient),
)
```