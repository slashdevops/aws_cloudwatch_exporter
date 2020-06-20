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

// https://aws.amazon.com/blogs/developer/mocking-out-then-aws-sdk-for-go-for-unit-testing/

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestNewSessionWithEnvVars(t *testing.T) {

	testCases := []struct {
		Name        string
		Description string
		EnvVars     map[string]string
		Expected    map[string]string
	}{
		{
			Name:        "UsingEnvVarsBasicAndToken",
			Description: "Using AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_SESSION_TOKEN Env Vars",
			EnvVars: map[string]string{
				"AWS_SHARED_CREDENTIALS_FILE": "/tmp/nothing", // This is very important to avoid the use of your own credentials file ~.aws/credentials
				"AWS_CONFIG_FILE":             "/tmp/nothing", // This is very important to avoid the use of your own config file ~.aws/config
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
			log.Debug(tc.Description)

			// Set the Environment Variables if exist
			for key, value := range tc.EnvVars {
				log.Debugf("Setting Env Var: %s", key)
				os.Setenv(key, value)
			}

			// Create the session with the arguments
			s := NewSession()

			// Get the result credentials and error
			c, err := s.Config.Credentials.Get()
			if err != nil {
				t.Errorf("The session creation fail with error: %s", err)
			}

			if tc.Expected["AWS_REGION"] != *s.Config.Region {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", *s.Config.Region, tc.Expected["AWS_REGION"])
			}

			if tc.Expected["AWS_ACCESS_KEY_ID"] != c.AccessKeyID {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.AccessKeyID, tc.Expected["AWS_ACCESS_KEY_ID"])
			}

			if tc.Expected["AWS_SECRET_ACCESS_KEY"] != c.SecretAccessKey {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SecretAccessKey, tc.Expected["AWS_SECRET_ACCESS_KEY"])
			}

			if tc.Expected["AWS_SESSION_TOKEN"] != c.SessionToken {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SessionToken, tc.Expected["AWS_SESSION_TOKEN"])
			}
		})
	}
}

func TestNewSessionWithFiles(t *testing.T) {

	testCases := []struct {
		Name        string
		Description string
		EnvVars     map[string]string
		Expected    map[string]string
	}{
		{
			Name:        "UsingProfileAndConfigStateFile",
			Description: "Using ",
			EnvVars: map[string]string{
				"AWS_SDK_LOAD_CONFIG":         "true",
				"AWS_SHARED_CREDENTIALS_FILE": "testdata/default/credentials", // This is very important to avoid the use of your own credentials file ~.aws/credentials
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":           "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY":       "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_REGION":                  "eu-west-1",
				"AWS_SHARED_CREDENTIALS_FILE": "EnvConfigCredentials",
			},
		},
		{
			Name:        "UsingProfileAndConfigFile",
			Description: "Using ",
			EnvVars: map[string]string{
				"AWS_SHARED_CREDENTIALS_FILE": "testdata/case1/credentials", // This is very important to avoid the use of your own credentials file ~.aws/credentials
				"AWS_CONFIG_FILE":             "testdata/case1/config",      // This is very important to avoid the use of your own credentials file ~.aws/credentials
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":           "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY":       "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_REGION":                  "eu-west-1",
				"AWS_SHARED_CREDENTIALS_FILE": "EnvConfigCredentials",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Using the same logger of the package aws
			log.Debug(tc.Description)

			// Set the Environment Variables if exist
			for key, value := range tc.EnvVars {
				log.Debugf("Setting Env Var: %s", key)
				os.Setenv(key, value)
			}

			// Create the session with the arguments
			s := NewSession()

			// Get the result credentials and error
			c, err := s.Config.Credentials.Get()
			if err != nil {
				t.Errorf("The session creation fail with error: %s", err)
			}

			if tc.Expected["AWS_REGION"] != *s.Config.Region {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", *s.Config.Region, tc.Expected["AWS_REGION"])
			}

			if tc.Expected["AWS_ACCESS_KEY_ID"] != c.AccessKeyID {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.AccessKeyID, tc.Expected["AWS_ACCESS_KEY_ID"])
			}

			if tc.Expected["AWS_SECRET_ACCESS_KEY"] != c.SecretAccessKey {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SecretAccessKey, tc.Expected["AWS_SECRET_ACCESS_KEY"])
			}

			if tc.Expected["AWS_SHARED_CREDENTIALS_FILE"] != c.ProviderName {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.ProviderName, tc.Expected["AWS_SHARED_CREDENTIALS_FILE"])
			}

		})
	}
}

func TestNewSessionWithConfigAndEnvVars(t *testing.T) {

	testCases := []struct {
		Name        string
		Description string
		EnvVars     map[string]string
		Expected    map[string]string
	}{
		{
			Name:        "UsingConfig",
			Description: "Using ",
			EnvVars: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_SESSION_TOKEN":     "ConfigToken",
			},
			Expected: map[string]string{
				"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
				"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"AWS_SESSION_TOKEN":     "ConfigToken",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Using the same logger of the package aws
			log.Debug(tc.Description)

			// Set the Environment Variables if exist
			for key, value := range tc.EnvVars {
				log.Debugf("Setting Env Var: %s", key)
				os.Setenv(key, value)
			}

			// Create the session with the arguments
			s := NewSession()

			// Get the result credentials and error
			c, err := s.Config.Credentials.Get()
			if err != nil {
				t.Errorf("The session creation fail with error: %s", err)
			}

			if tc.Expected["AWS_ACCESS_KEY_ID"] != c.AccessKeyID {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.AccessKeyID, tc.Expected["AWS_ACCESS_KEY_ID"])
			}

			if tc.Expected["AWS_SECRET_ACCESS_KEY"] != c.SecretAccessKey {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SecretAccessKey, tc.Expected["AWS_SECRET_ACCESS_KEY"])
			}

			if tc.Expected["AWS_SESSION_TOKEN"] != c.SessionToken {
				t.Errorf("\n\t Gotten: %s \n\t Expected: %s", c.SessionToken, tc.Expected["AWS_SESSION_TOKEN"])
			}
		})
	}
}
