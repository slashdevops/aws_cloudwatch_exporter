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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)

// https://docs.Credentials.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
// https://docs.Credentials.amazon.com/sdk-for-go/api/aws/session/
func NewSession() *session.Session {

	awsConf := aws.Config{CredentialsChainVerboseErrors: aws.Bool(true)}
	awsSessOpts := session.Options{Config: awsConf}

	// in case of assume role with MFA
	awsSessOpts.AssumeRoleTokenProvider = stscreds.StdinTokenProvider

	if len(os.Getenv("AWS_ACCESS_KEY_ID")) == 0 ||
		len(os.Getenv("AWS_SDK_LOAD_CONFIG")) > 1 ||
		len(os.Getenv("AWS_SHARED_CREDENTIALS_FILE")) > 1 ||
		len(os.Getenv("AWS_CONFIG_FILE")) > 1 {
		awsSessOpts.SharedConfigState = session.SharedConfigEnable
	}

	if len(os.Getenv("AWS_REGION")) > 0 {
		log.Debugf("Using AWS Region: %s", os.Getenv("AWS_REGION"))
		awsConf.Region = aws.String(os.Getenv("AWS_REGION"))
	}

	if len(os.Getenv("AWS_PROFILE")) > 0 {
		log.Debugf("Using AWS Profile: %s", os.Getenv("AWS_PROFILE"))
		awsSessOpts.Profile = os.Getenv("AWS_PROFILE")
	}

	// Try default credential provider chain
	// awsSession, err := session.NewSession(&awsConf)
	awsSession, err := session.NewSessionWithOptions(awsSessOpts)
	if err != nil {
		log.Fatalf("Failed to create AWS Session using default chain:%s", err)
	}

	if len(os.Getenv("AWS_ROLE_ARN")) > 0 {
		log.Debugf("Create new session assuming role: %s", os.Getenv("AWS_ROLE_ARN"))
		// Override session AWS with the new credentials provided after assume the role
		awsConf.Credentials = stscreds.NewCredentials(awsSession, os.Getenv("AWS_ROLE_ARN"))
		awsSession = session.Must(session.NewSession(&awsConf))
	}

	return awsSession
}
