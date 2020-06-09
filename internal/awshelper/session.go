package awshelper

import (
	"os"
	"reflect"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)

// https://docs.Credentials.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
// https://docs.Credentials.amazon.com/sdk-for-go/api/aws/session/
func NewSession(c *config.AWS) *session.Session {
	awsConf := aws.Config{CredentialsChainVerboseErrors: aws.Bool(true)}
	awsSessOpts := session.Options{Config: awsConf}
	awsSession := &session.Session{}
	awsSession.Config = &awsConf

	// Case 1: When config.AWS structure is empty, use default credentials chain providers
	// Trying to use default credential provider chain to find AWS credentials
	// 1. Environment variables.
	// 2. Shared credentials file.
	// 3. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.
	// 4. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.
	// https://docs.Credentials.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
	// https://docs.Credentials.amazon.com/sdk-for-go/api/aws/defaults/#CredChain
	if reflect.DeepEqual(&config.AWS{}, c) {
		log.Debug("Creating AWS Session from default credential provider chain")

		// force the use of standar files ~/.aws/credentials and ~/.aws/config
		// The same export AWS_SDK_LOAD_CONFIG="true"
		// this force to use AWS_PROFILE from both files, and not only from ~/.aws/credentials
		// like aws cli when use --profile
		if (os.Getenv("AWS_SHARED_CREDENTIALS_FILE") == "") && (os.Getenv("AWS_CONFIG_FILE") == "") {
			awsSessOpts.SharedConfigState = session.SharedConfigEnable
		}

		// When mfa_serial is enabled in the profile the session fail with the message:
		// panic: AssumeRoleTokenProviderNotSetError: assume role with MFA enabled, but AssumeRoleTokenProvider session option not set.
		// this is to avoid it.
		// This force to ask you for the token
		// NOTE: Until now doesn't exist anye env var to define the tocken provider (AssumeRoleTokenProvider)
		if os.Getenv("AWS_PROFILE") != "" {
			_, err := session.NewSessionWithOptions(awsSessOpts)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case "AssumeRoleTokenProviderNotSetError":
						log.Infof("The profile '%s' is using MFA configuration", os.Getenv("AWS_PROFILE"))
						awsSessOpts.AssumeRoleTokenProvider = stscreds.StdinTokenProvider
					default:
						log.Error(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					log.Error(err.Error())
				}
			}
		}

		awsSession = session.Must(session.NewSessionWithOptions(awsSessOpts))
	}

	// Case 2: When config.AWS structure is not empty which is mean use local config file "aws" section
	if !reflect.DeepEqual(&config.AWS{}, c) {
		log.Debug("Creating AWS Session from config.AWS")

		if c.Region != "" {
			log.Debugf("Using AWS Region: %s", c.Region)
			awsConf.Region = aws.String(c.Region)
		}

		// profile exist, necessary to use .aws/credentials and .aws/config wherever they are
		if c.Profile != "" && !c.SharedConfigState {
			c.SharedConfigState = true
			awsSessOpts.Profile = c.Profile
		}

		// Case 2.1: When this is enabled all the credentials where load from files, not from config
		// Force to use .aws/credentials and .aws/config or custom location
		if c.SharedConfigState {
			log.Debug("Enabling Session SharedConfigState")
			awsSessOpts.SharedConfigState = session.SharedConfigEnable

			// Case 2.1.1: Different path for credentials/config
			if len(c.SharedCredentialsFile) > 0 {
				log.Debugf("Using custom credential files: %s", c.SharedCredentialsFile)

				// Using custom files and location
				if len(c.ConfigFile) > 0 {
					log.Debugf("Using custom config files: %s", c.ConfigFile)
					files := append(c.SharedCredentialsFile, c.ConfigFile...)
					log.Debugf("Using custom credential/config files: %s", files)
					awsSessOpts.SharedConfigFiles = files
				} else {
					awsSessOpts.SharedConfigFiles = c.SharedCredentialsFile
				}
			}

			// When mfa_serial is enabled in the profile the session fail with the message:
			// panic: AssumeRoleTokenProviderNotSetError: assume role with MFA enabled, but AssumeRoleTokenProvider session option not set.
			// this is to avoid it.
			// This force to ask you for the token
			// NOTE: Until now doesn't exist anye env var to define the tocken provider (AssumeRoleTokenProvider)
			_, err := session.NewSessionWithOptions(awsSessOpts)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case "AssumeRoleTokenProviderNotSetError":
						log.Infof("The profile '%s' is using MFA configuration", c.Profile)
						awsSessOpts.AssumeRoleTokenProvider = stscreds.StdinTokenProvider
					default:
						log.Error(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					log.Error(err.Error())
				}
			}

			awsSession = session.Must(session.NewSessionWithOptions(awsSessOpts))

		} else { // Case 2.2: When exist access_key, secret_key profile,etc in conf file region come from conf file also

			if c.Region != "" {
				log.Debugf("Using AWS Region: %s", c.Region)
				awsConf.Region = aws.String(c.Region)
			}

			log.Debug("Using AWS AccessKey, SecretKey and SessionToken")
			awsConf.Credentials = credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, c.SessionToken)
			awsSession = session.Must(session.NewSession(&awsConf))

			// Case 2.2.1: AWS and role provided, you must assume the role and create new session
			// using this role
			if c.RoleArn != "" {
				log.Debugf("Assuming a Role: %s", c.RoleArn)

				// Override session AWS with the new credentials provided after assume the role
				awsConf.Credentials = stscreds.NewCredentials(awsSession, c.RoleArn)
				awsSession = session.Must(session.NewSession(&awsConf))
			}
		}
	}
	return awsSession
}
