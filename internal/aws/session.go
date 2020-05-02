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

// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
// https://docs.aws.amazon.com/sdk-for-go/api/aws/session/

func NewSession(c *config.Server) (*session.Session, error) {
	awsConf := aws.Config{CredentialsChainVerboseErrors: aws.Bool(true)}
	awsSession := &session.Session{}
	awsSessionOptions := session.Options{}

	// Case 1: When config.AWS structure is empty, use default credentials chain providers
	// Trying to use default credential provider chain to find AWS credentials
	// 1. Environment variables.
	// 2. Shared credentials file.
	// 3. If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.
	// 4. If your application uses an ECS task definition or RunTask API operation, IAM role for tasks.
	// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials
	// https://docs.aws.amazon.com/sdk-for-go/api/aws/defaults/#CredChain
	if reflect.DeepEqual(config.AWS{}, c.AWS) {
		c.Logger.Debug("Creating AWS Session from default credential provider chain")

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
	if !reflect.DeepEqual(config.AWS{}, c.AWS) {
		c.Logger.Debug("Creating AWS Session from config.AWS")

		// profile exist, necessary to use .aws/credentials and .aws/config wherever they are
		if c.AWS.Profile != "" && !c.AWS.SharedConfigState {
			c.AWS.SharedConfigState = true
			awsSessionOptions.Profile = c.AWS.Profile
		}

		// Case 2.1: When this is enabled all the credentials where load from files, not from config
		// Force to use .aws/credentials and .aws/config or custom location
		if c.AWS.SharedConfigState {
			c.Logger.Debug("Enabling Session SharedConfigState")
			awsSessionOptions.SharedConfigState = session.SharedConfigEnable

			// Case 2.1.1: Different path for credentials/config
			if len(c.AWS.CredentialsFile) > 0 {
				c.Logger.Debugf("Using custom credential files: %s", c.AWS.CredentialsFile)
				if len(c.AWS.ConfigFile) > 0 {
					c.Logger.Debugf("Using custom config files: %s", c.AWS.ConfigFile)
					files := append(c.AWS.CredentialsFile, c.AWS.ConfigFile...)
					c.Logger.Debugf("Using custom credential/config files: %s", files)
					awsSessionOptions.SharedConfigFiles = files
				} else {
					awsSessionOptions.SharedConfigFiles = c.AWS.CredentialsFile
				}
			}

			// Try the new session, if fail control the errors
			awsSession, err := session.NewSessionWithOptions(awsSessionOptions)
			if err != nil {
				if aerr, ok := err.(awserr.Error); ok {
					switch aerr.Code() {
					case "AssumeRoleTokenProviderNotSetError":
						c.Logger.Infof("The profile '%s' is using MFA configuration", c.AWS.Profile)
						awsSessionOptions.AssumeRoleTokenProvider = stscreds.StdinTokenProvider
					default:
						c.Logger.Println(aerr.Error())
					}
				} else {
					// Print the error, cast err to awserr.Error to get the Code and
					// Message from an error.
					c.Logger.Println(err.Error())
				}
			}

			awsSession = session.Must(session.NewSessionWithOptions(awsSessionOptions))
			return awsSession, nil

		} else { // Case 2.2: When exist access_key, secret_key profile,etc in conf file
			// region come from conf file also
			if c.AWS.Region != "" {
				c.Logger.Debugf("Using AWS Region: %s", c.AWS.Region)
				awsConf.Region = aws.String(c.AWS.Region)
			}

			c.Logger.Debug("Using AWS AccessKey, SecretKey and SessionToken")
			awsConf.Credentials = credentials.NewStaticCredentials(c.AWS.AccessKeyID, c.AWS.SecretAccessKey, c.AWS.SessionToken)
			awsSession = session.Must(session.NewSession(&awsConf))

			// Case 2.2.1: AWS and role provided, you must assume the role and create new session
			// using this role
			if c.AWS.RoleArn != "" {
				c.Logger.Debugf("Assuming a Role: %s", c.AWS.RoleArn)

				// Override session AWS with the new credentials provided after assume the role
				awsConf.Credentials = stscreds.NewCredentials(awsSession, c.AWS.RoleArn)
				// Override session with the new session after assume a role and get new credentials
				awsSession = session.Must(session.NewSession(&awsConf))
			}
			return awsSession, nil
		}
	}
	return awsSession, nil
}
