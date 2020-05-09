package aws

import (
	"os"
	"reflect"

	"github.com/prometheus/common/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// https://docs.Credentials.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
// https://docs.Credentials.amazon.com/sdk-for-go/api/aws/session/

func NewSession(c *config.Credentials) (*session.Session, error) {
	awsConf := aws.Config{CredentialsChainVerboseErrors: aws.Bool(true)}
	awsSession := &session.Session{}
	awsSessionOptions := session.Options{}

	// Case 1: When config.AWS structure is empty, use default credentials chain providers
	// Trying to use default credential provider chain to find AWS credentials
	// 1. Environment variables.
	// 2. Shared credentials file.
	// 3. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.
	// 4. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.
	// https://docs.Credentials.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
	// https://docs.Credentials.amazon.com/sdk-for-go/api/aws/defaults/#CredChain
	if reflect.DeepEqual(&config.Credentials{}, c) {
		log.Debug("Creating AWS Session from default credential provider chain")

		// Case 1.1: It is Defined the AWS_PROFILE env var, so this read the credentials from .aws/credentials file
		// using the [profile profileName] where profileName = AWS_PROFILE
		if os.Getenv("AWS_PROFILE") != "" {

			// CASE 1.1.1: The credentials files .aws/credentials is in a custom location
			if os.Getenv("AWS_SHARED_CREDENTIALS_FILE") != "" {
				if os.Getenv("AWS_CONFIG_FILE") != "" {
					awsSessionOptions.SharedConfigFiles = []string{
						os.Getenv("AWS_SHARED_CREDENTIALS_FILE"),
						os.Getenv("AWS_CONFIG_FILE"),
					}
					awsSessionOptions.Profile = os.Getenv("AWS_PROFILE")
					awsSession = session.Must(session.NewSessionWithOptions(awsSessionOptions))
					return awsSession, nil
				} else {
					awsConf.Credentials = credentials.NewSharedCredentials(os.Getenv("AWS_SHARED_CREDENTIALS_FILE"), os.Getenv("AWS_PROFILE"))
				}

			} else {
				awsConf.Credentials = credentials.NewSharedCredentials("", os.Getenv("AWS_PROFILE"))
			}
			awsSession = session.Must(session.NewSession(&awsConf))
			return awsSession, nil
		}

		// Case 1.2: TBI, When in necessary assume a role

		// Use default chain
		awsSession = session.Must(session.NewSession())
		return awsSession, nil
	}

	// Case 2: When config.AWS structure is not empty which is mean use local config file "aws" section
	if !reflect.DeepEqual(&config.Credentials{}, c) {
		log.Debug("Creating AWS Session from config.AWS")

		// profile exist, necessary to use .aws/credentials and .aws/config wherever they are
		if c.Profile != "" && !c.SharedConfigState {
			c.SharedConfigState = true
			awsSessionOptions.Profile = c.Profile
		}

		// Case 2.1: When this is enabled all the credentials where load from files, not from config
		// Force to use .aws/credentials and .aws/config or custom location
		if c.SharedConfigState {
			log.Debug("Enabling Session SharedConfigState")
			awsSessionOptions.SharedConfigState = session.SharedConfigEnable

			// Case 2.1.1: Different path for credentials/config
			if len(c.CredentialsFile) > 0 {
				log.Debugf("Using custom credential files: %s", c.CredentialsFile)
				if len(c.ConfigFile) > 0 {
					log.Debugf("Using custom config files: %s", c.ConfigFile)
					files := append(c.CredentialsFile, c.ConfigFile...)
					log.Debugf("Using custom credential/config files: %s", files)
					awsSessionOptions.SharedConfigFiles = files
				} else {
					awsSessionOptions.SharedConfigFiles = c.CredentialsFile
				}
			}

			// Try the new session, if fail control the errors
			awsSession, err := session.NewSessionWithOptions(awsSessionOptions)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case "AssumeRoleTokenProviderNotSetError":
						log.Infof("The profile '%s' is using MFA configuration", c.Profile)
						awsSessionOptions.AssumeRoleTokenProvider = stscreds.StdinTokenProvider
					default:
						log.Error(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					log.Error(err.Error())
				}
			}

			awsSession = session.Must(session.NewSessionWithOptions(awsSessionOptions))
			return awsSession, nil

		} else { // Case 2.2: When exist access_key, secret_key profile,etc in conf file
			// region come from conf file also
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
				// Override session with the new session after assume a role and get new credentials
				awsSession = session.Must(session.NewSession(&awsConf))
			}
			return awsSession, nil
		}
	}
	return awsSession, nil
}
