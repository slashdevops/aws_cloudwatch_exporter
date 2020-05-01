package aws

// https://aws.amazon.com/blogs/developer/mocking-out-then-aws-sdk-for-go-for-unit-testing/

import (
	"os"
	"testing"

	"github.com/aimroot/aws_cloudwatch_exporter/config"

	"github.com/sirupsen/logrus"
)

const (
	LogLevel = logrus.DebugLevel
)

func TestNewSessionWithEnvVars(t *testing.T) {

	testCases := []struct {
		Name        string
		Description string
		Args        *config.Server
		EnvVars     map[string]string
		Expected    map[string]string
	}{
		{
			Name:        "UsingEnvVarsBasicAndToken",
			Description: "Using AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_SESSION_TOKEN Env Vars",
			Args: &config.Server{
				AWS:    config.AWS{},
				Logger: logrus.New(),
			},
			EnvVars: map[string]string{
				"AWS_SHARED_CREDENTIALS_FILE": "/tmp/nothing", // This is very important to avoid the use of your own credentials
				"AWS_CONFIG_FILE":             "/tmp/nothing", // This is very important to avoid the use of your own credentials
				"AWS_ACCESS_KEY_ID":           "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY":       "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_REGION":                  "eu-west-1",
				"AWS_SESSION_TOKEN":           "TheToken",
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_REGION":            "eu-west-1",
				"AWS_SESSION_TOKEN":     "TheToken",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Using the same logger of the package aws
			tc.Args.Logger.SetLevel(LogLevel)
			tc.Args.Logger.SetFormatter(&logrus.JSONFormatter{})
			tc.Args.Logger.Debug(tc.Description)

			// Set the Environment Variables if exist
			for key, value := range tc.EnvVars {
				tc.Args.Logger.Debugf("Setting Env Var: %s", key)
				os.Setenv(key, value)
			}

			// Create the session with the arguments
			s, _ := NewSession(tc.Args)

			// Get the result credentials and error
			c, err := s.Config.Credentials.Get()
			if err != nil {
				t.Errorf("The session creation fail with error: %s", err)
			}

			for key, value := range tc.Expected {

				if key == "AWS_REGION" && *s.Config.Region != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", *s.Config.Region, value)
				}
				if key == "AWS_ACCESS_KEY_ID" && c.AccessKeyID != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.AccessKeyID, value)
				}
				if key == "AWS_SECRET_ACCESS_KEY" && c.SecretAccessKey != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SecretAccessKey, value)
				}
				if key == "AWS_SESSION_TOKEN" && c.SessionToken != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SessionToken, value)
				}
			}
		})
	}
}

func TestNewSessionWithFiles(t *testing.T) {

	testCases := []struct {
		Name        string
		Description string
		Args        *config.Server
		Expected    map[string]string
	}{
		{
			Name:        "UsingProfileAndConfigStateFile",
			Description: "Using ",
			Args: &config.Server{
				AWS: config.AWS{
					Profile:           "default",
					SharedConfigState: true,
					CredentialsFile:   []string{"testdata/default/credentials", "testdata/default/config"},
				},
				Logger: logrus.New(),
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_REGION":            "eu-west-1",
				"AWS_SESSION_TOKEN":     "DefaultToken",
			},
		},
		{
			Name:        "UsingProfileAndConfigStateFile",
			Description: "Using cas1 profile from testdata",
			Args: &config.Server{
				AWS: config.AWS{
					Profile:           "case1",
					SharedConfigState: true,
					CredentialsFile:   []string{"testdata/case1/credentials", "testdata/case1/config"},
				},
				Logger: logrus.New(),
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_REGION":            "us-west-2",
				"AWS_SESSION_TOKEN":     "Case1Token",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Using the same logger of the package aws
			tc.Args.Logger.SetLevel(LogLevel)
			tc.Args.Logger.SetFormatter(&logrus.JSONFormatter{})
			tc.Args.Logger.Debug(tc.Description)

			// Create the session with the arguments
			s, _ := NewSession(tc.Args)

			// Get the result credentials and error
			c, err := s.Config.Credentials.Get()
			if err != nil {
				t.Errorf("The session creation fail with error: %s", err)
			}

			for key, value := range tc.Expected {

				if key == "AWS_REGION" && *s.Config.Region != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", *s.Config.Region, value)
				}
				if key == "AWS_ACCESS_KEY_ID" && c.AccessKeyID != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.AccessKeyID, value)
				}
				if key == "AWS_SECRET_ACCESS_KEY" && c.SecretAccessKey != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SecretAccessKey, value)
				}
				if key == "AWS_SESSION_TOKEN" && c.SessionToken != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SessionToken, value)
				}
			}
		})
	}
}

func TestNewSessionWithConfig(t *testing.T) {

	testCases := []struct {
		Name        string
		Description string
		Args        *config.Server
		Expected    map[string]string
	}{
		{
			Name:        "UsingConfig",
			Description: "Using ",
			Args: &config.Server{
				AWS: config.AWS{
					AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
					SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					//Region:          "us-west-2",
					SessionToken: "ConfigToken",
				},
				Logger: logrus.New(),
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				//"AWS_REGION":            "us-west-2",
				"AWS_SESSION_TOKEN": "ConfigToken",
			},
		},
		{
			Name:        "UsingConfig",
			Description: "Using cas1 profile from testdata",
			Args: &config.Server{
				AWS: config.AWS{
					AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
					SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					Region:          "us-west-2",
					//RoleArn:         "arn:aws:iam::123456789012:role/role-name",
					SessionToken: "ConfigToken",
				},
				Logger: logrus.New(),
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_REGION":            "us-west-2",
				"AWS_SESSION_TOKEN":     "ConfigToken",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Using the same logger of the package aws
			tc.Args.Logger.SetLevel(LogLevel)
			tc.Args.Logger.SetFormatter(&logrus.JSONFormatter{})
			tc.Args.Logger.Debug(tc.Description)

			// Create the session with the arguments
			s, _ := NewSession(tc.Args)

			// Get the result credentials and error
			c, err := s.Config.Credentials.Get()
			if err != nil {
				t.Errorf("The session creation fail with error: %s", err)
			}

			for key, value := range tc.Expected {

				if key == "AWS_REGION" && *s.Config.Region != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", *s.Config.Region, value)
				}
				if key == "AWS_ACCESS_KEY_ID" && c.AccessKeyID != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.AccessKeyID, value)
				}
				if key == "AWS_SECRET_ACCESS_KEY" && c.SecretAccessKey != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SecretAccessKey, value)
				}
				if key == "AWS_SESSION_TOKEN" && c.SessionToken != value {
					t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SessionToken, value)
				}
			}
		})
	}
}
