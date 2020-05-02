package aws

import (
	"os"
	"reflect"

	"github.com/slashdevops/aws_cloudwatch_exporter/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// https://docs.Credentials.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
// https://docs.Credentials.amazon.com/sdk-for-go/api/aws/session/

func NewSession(c *config.All) (*session.Session, error) {
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
	if reflect.DeepEqual(config.CredentialsConf{}, c.Credentials) {
		c.Application.Logger.Debug("Creating AWS Session from default credential provider chain")

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
	if !reflect.DeepEqual(config.CredentialsConf{}, c.Credentials) {
		c.Application.Logger.Debug("Creating AWS Session from config.AWS")

		// profile exist, necessary to use .aws/credentials and .aws/config wherever they are
		if c.Credentials.Profile != "" && !c.Credentials.SharedConfigState {
			c.Credentials.SharedConfigState = true
			awsSessionOptions.Profile = c.Credentials.Profile
		}

		// Case 2.1: When this is enabled all the credentials where load from files, not from config
		// Force to use .aws/credentials and .aws/config or custom location
		if c.Credentials.SharedConfigState {
			c.Application.Logger.Debug("Enabling Session SharedConfigState")
			awsSessionOptions.SharedConfigState = session.SharedConfigEnable

			// Case 2.1.1: Different path for credentials/config
			if len(c.Credentials.CredentialsFile) > 0 {
				c.Application.Logger.Debugf("Using custom credential files: %s", c.Credentials.CredentialsFile)
				if len(c.Credentials.ConfigFile) > 0 {
					c.Application.Logger.Debugf("Using custom config files: %s", c.Credentials.ConfigFile)
					files := append(c.Credentials.CredentialsFile, c.Credentials.ConfigFile...)
					c.Application.Logger.Debugf("Using custom credential/config files: %s", files)
					awsSessionOptions.SharedConfigFiles = files
				} else {
					awsSessionOptions.SharedConfigFiles = c.Credentials.CredentialsFile
				}
			}

			// Try the new session, if fail control the errors
			awsSession, err := session.NewSessionWithOptions(awsSessionOptions)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case "AssumeRoleTokenProviderNotSetError":
						c.Application.Logger.Infof("The profile '%s' is using MFA configuration", c.Credentials.Profile)
						awsSessionOptions.AssumeRoleTokenProvider = stscreds.StdinTokenProvider
					default:
						c.Application.Logger.Println(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					c.Application.Logger.Println(err.Error())
				}
			}

			awsSession = session.Must(session.NewSessionWithOptions(awsSessionOptions))
			return awsSession, nil

		} else { // Case 2.2: When exist access_key, secret_key profile,etc in conf file
			// region come from conf file also
			if c.Credentials.Region != "" {
				c.Application.Logger.Debugf("Using AWS Region: %s", c.Credentials.Region)
				awsConf.Region = aws.String(c.Credentials.Region)
			}

			c.Application.Logger.Debug("Using AWS AccessKey, SecretKey and SessionToken")
			awsConf.Credentials = credentials.NewStaticCredentials(c.Credentials.AccessKeyID, c.Credentials.SecretAccessKey, c.Credentials.SessionToken)
			awsSession = session.Must(session.NewSession(&awsConf))

			// Case 2.2.1: AWS and role provided, you must assume the role and create new session
			// using this role
			if c.Credentials.RoleArn != "" {
				c.Application.Logger.Debugf("Assuming a Role: %s", c.Credentials.RoleArn)

				// Override session AWS with the new credentials provided after assume the role
				awsConf.Credentials = stscreds.NewCredentials(awsSession, c.Credentials.RoleArn)
				// Override session with the new session after assume a role and get new credentials
				awsSession = session.Must(session.NewSession(&awsConf))
			}
			return awsSession, nil
		}
	}
	return awsSession, nil
}
