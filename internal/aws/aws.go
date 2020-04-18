package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aimroot/aws_cloudwatch_exporter/config"
	"github.com/aws/aws-sdk-go/aws/session"
)

type AWSSession struct {
	conf   *config.Config
	logger *log.Logger
}

func NewAWSSession(l *log.Logger, c *config.Config) *AWSSession {

	if c.AWS.Credentials != nil{

	}
	aws_config := &aws.Config{Region: }
	sess := session.Must(session.NewSession())

}
