/*
Copyright © 2020 Christian González Di Antonio christian@slashdevops.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package awshelper

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)

// https://docs.Credentials.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
// https://docs.Credentials.amazon.com/sdk-for-go/api/aws/session/
func NewSession(c *config.AWS) *session.Session {

	awsConf := aws.Config{CredentialsChainVerboseErrors: aws.Bool(true)}
	awsSessOpts := session.Options{Config: awsConf}

	// in case of assume role with MFA
	awsSessOpts.AssumeRoleTokenProvider = stscreds.StdinTokenProvider

	if len(os.Getenv("AWS_ACCESS_KEY_ID")) == 0 {
		awsSessOpts.SharedConfigState = session.SharedConfigEnable
	}

	if len(c.Region) > 0 {
		log.Debugf("Using AWS Region: %s", c.Region)
		awsConf.Region = aws.String(c.Region)
	}

	if c.Profile != "" {
		awsSessOpts.Profile = c.Profile
		log.Debug("Enabling Session SharedConfigState")
	}

	// Try default credential provider chain
	// awsSession, err := session.NewSession(&awsConf)
	awsSession, err := session.NewSessionWithOptions(awsSessOpts)
	if err != nil {
		log.Fatalf("Failed to create AWS Session using default chain:%s", err)
	}

	if len(c.RoleArn) > 0 {
		log.Debugf("Create new session assuming role: %s", c.RoleArn)
		// Override session AWS with the new credentials provided after assume the role
		awsConf.Credentials = stscreds.NewCredentials(awsSession, c.RoleArn)
		awsSession = session.Must(session.NewSession(&awsConf))
	}

	return awsSession
}
