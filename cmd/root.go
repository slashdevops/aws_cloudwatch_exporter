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
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const (
	appName             = "aws_cloudwatch_exporter"
	appDescription      = `This is an AWS CloudWatch exporter for prometheus.io`
	appDescriptionShort = "AWS CloudWatch exporter for prometheus.io"
	appGitRepository    = "https://github.com/slashdevops/aws_cloudwatch_exporter"
	appMetricsPath      = "/metrics"
	appHealthPath       = "/health"
)

// rootCmd represents the base command when called without any subcommands
var (
	// Global vars
	conf config.All

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: appDescriptionShort,
		Long:  appDescription,
	}
	log = logrus.New()
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Debug
	rootCmd.PersistentFlags().BoolVar(&conf.Server.Debug, "debug", false, "Enable debug messages in logs")
	if err := viper.BindPFlag("server.debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		log.Error(err)
	}

	// Files
	rootCmd.PersistentFlags().StringVar(&conf.Application.ServerFile, "serverFile", "server.yaml", "The file with the server configuration, see: README.md")
	if err := viper.BindPFlag("application.serverFile", rootCmd.PersistentFlags().Lookup("serverFile")); err != nil {
		log.Error(err)
	}

	rootCmd.PersistentFlags().StringVar(&conf.Application.CredentialsFile, "credentialsFile", "credentials.yaml", "The file with the AWS Credentials configuration")
	if err := viper.BindPFlag("application.credentialsFile", rootCmd.PersistentFlags().Lookup("credentialsFile")); err != nil {
		log.Error(err)
	}

	rootCmd.PersistentFlags().StringSliceVar(&conf.Application.MetricsFiles, "metricsFiles", nil, "Metrics files, example: --metricsFiles ~/tmp/queries/m1.yaml --metricsFiles ~/tmp/queries/m2.yml")
	if err := viper.BindPFlag("application.metricsFiles", rootCmd.PersistentFlags().Lookup("metricsFiles")); err != nil {
		log.Error(err)
	}

	// AWS Credentials conf
	// AccessKeyID
	rootCmd.PersistentFlags().StringVar(&conf.AWS.AccessKeyID, "access_key_id", "", "The AWS AccessKeyID, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.access_key_id", rootCmd.PersistentFlags().Lookup("access_key_id")); err != nil {
		log.Error(err)
	}

	// SecretAccessKey
	rootCmd.PersistentFlags().StringVar(&conf.AWS.SecretAccessKey, "secret_access_key", "", "The AWS SecretAccessKey, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.secret_access_key", rootCmd.PersistentFlags().Lookup("secret_access_key")); err != nil {
		log.Error(err)
	}

	// SessionToken
	rootCmd.PersistentFlags().StringVar(&conf.AWS.SessionToken, "session_token", "", "The AWS SessionToken, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.session_token", rootCmd.PersistentFlags().Lookup("session_token")); err != nil {
		log.Error(err)
	}

	// Region
	rootCmd.PersistentFlags().StringVar(&conf.AWS.Region, "region", "", "The AWS Region, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.region", rootCmd.PersistentFlags().Lookup("region")); err != nil {
		log.Error(err)
	}

	// Profile
	rootCmd.PersistentFlags().StringVar(&conf.AWS.Profile, "profile", "", "The AWS profile name defined in .aws/config or .aws/credential")
	if err := viper.BindPFlag("aws.profile", rootCmd.PersistentFlags().Lookup("profile")); err != nil {
		log.Error(err)
	}

	// RoleArn
	rootCmd.PersistentFlags().StringVar(&conf.AWS.RoleArn, "role_arn", "", "The AWS RoleArn, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.role_arn", rootCmd.PersistentFlags().Lookup("role_arn")); err != nil {
		log.Error(err)
	}

	// RoleSessionName
	rootCmd.PersistentFlags().StringVar(&conf.AWS.RoleSessionName, "role_session_name", "", "The AWS RoleSessionName, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.role_session_name", rootCmd.PersistentFlags().Lookup("role_session_name")); err != nil {
		log.Error(err)
	}

	// WebIdentityTokenFile
	rootCmd.PersistentFlags().StringVar(&conf.AWS.WebIdentityTokenFile, "web_identity_token_file", "", "The AWS WebIdentityTokenFile, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.web_identity_token_file", rootCmd.PersistentFlags().Lookup("web_identity_token_file")); err != nil {
		log.Error(err)
	}

	// ExternalID
	rootCmd.PersistentFlags().StringVar(&conf.AWS.ExternalID, "external_id", "", "The AWS ExternalID, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.external_id", rootCmd.PersistentFlags().Lookup("external_id")); err != nil {
		log.Error(err)
	}

	// MFASerial
	rootCmd.PersistentFlags().StringVar(&conf.AWS.MFASerial, "mfa_serial", "", "The AWS MFASerial, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.mfa_serial", rootCmd.PersistentFlags().Lookup("mfa_serial")); err != nil {
		log.Error(err)
	}

	// SharedConfigState
	rootCmd.PersistentFlags().BoolVar(&conf.AWS.SharedConfigState, "shared_config_state", true, "The AWS SharedConfigState, see: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html")
	if err := viper.BindPFlag("aws.shared_config_state", rootCmd.PersistentFlags().Lookup("shared_config_state")); err != nil {
		log.Error(err)
	}

	// SharedCredentialsFile
	rootCmd.PersistentFlags().StringSliceVar(&conf.AWS.SharedCredentialsFile, "shared_credential_file", nil, "The AWS SharedCredentialsFile, example: --shared_credential_file ~/.aws/credentials --shared_credential_file /etc/aws/credentials")
	if err := viper.BindPFlag("aws.shared_credential_file", rootCmd.PersistentFlags().Lookup("shared_credential_file")); err != nil {
		log.Error(err)
	}

	// ConfigFile
	rootCmd.PersistentFlags().StringSliceVar(&conf.AWS.ConfigFile, "config_file", nil, "The AWS ConfigFile, example: --config_file ~/.aws/config --config_file /etc/aws/config")
	if err := viper.BindPFlag("aws.config_file", rootCmd.PersistentFlags().Lookup("config_file")); err != nil {
		log.Error(err)
	}
}

func initConfig() {
	if conf.Server.LogFormat == "json" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{})
	}

	if conf.Server.Debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetOutput(os.Stdout)

	// Set the output of the message for the current logrus instance,
	// Output of logrus instance can be set to any io.writer
	log.Out = os.Stdout

	// Set default values
	conf.Application.Name = appName
	conf.Application.Description = appDescription
	conf.Application.GitRepository = appGitRepository
	conf.Application.MetricsPath = appMetricsPath
	conf.Application.HealthPath = appHealthPath
	conf.Application.Version = version.Version
	conf.Application.Revision = version.Revision
	conf.Application.GoVersion = version.GoVersion
	conf.Application.BuildUser = version.BuildUser
	conf.Application.BuildDate = version.BuildDate
	conf.Application.VersionInfo = version.Info()
	conf.Application.BuildInfo = version.BuildContext()
}

// this will be used for every commands that needs conf in files
func ReadAndValidateConfFromFiles() {
	parseConfigFiles(&conf)
	parseMetricsFiles(&conf)
	validateMetricsQueries(&conf)

	// expose all the configuration, just to check
	if conf.Server.Debug {
		log.Debug(conf.ToJSON())
		// log.VersionInfo(conf.ToYAML())
	}
}

// Unmarshall Yaml files into c config structure
func parseConfigFiles(c *config.All) {
	// Config files to be load
	files := []string{
		c.Application.ServerFile,
		c.Application.CredentialsFile,
	}

	for _, file := range files {
		log.Infof("Reading configuration file: %s", file)

		// fileNameNoExt := strings.TrimSuffix(file, filepath.Ext(file))

		log.Debugf("Parsing configuration file path: %s", file)
		log.Debugf("File: %s", filepath.Base(file))
		// log.Debugf("File without ext: %s", fileNameNoExt)
		log.Debugf("Location: %s", filepath.Dir(file))
		log.Debugf("File ext: %s", filepath.Ext(file)[1:])

		// viper.SetConfigName(fileNameNoExt)
		viper.SetConfigName(filepath.Base(file))
		viper.AddConfigPath(filepath.Dir(file))
		viper.SetConfigType(filepath.Ext(file)[1:])

		// Read env vars equals as the mapstructure is defined into the config.go
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		log.Debugf("Loading configuration from file: %s", file)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file, %s", err)
		}

		log.Debugf("Filling configuration structure from file: %s", file)
		err := viper.Unmarshal(&c)
		if err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
	}
}

// Unmarshall Yaml files into c config structure
// NOTE: Unfortunately viper.MergeInConfig() do the merge using override, so
// this is the reason to do not user it.
func parseMetricsFiles(c *config.All) {

	metricsQueries := MergeMetricsFiles(c.Application.MetricsFiles)

	if err := viper.MergeConfigMap(metricsQueries); err != nil {
		log.Errorf("Error merging MetricsQueries read from files into config structure: %s", err.Error())
	}

	if len(c.Application.MetricsFiles) > 0 {
		log.Debugf("Filling configuration structure from file: %s", c.Application.MetricsFiles)
		err := viper.Unmarshal(&c)
		if err != nil {
			log.Errorf("Unable to unmarshal viper config into config struct, %s", err.Error())
		}
	} else {
		log.Errorf("Metrics configuration file: \"%v\" doesn't exist", c.Application.MetricsFiles)
	}
}

// This function merge files with metrics queries into a map without override keys
func MergeMetricsFiles(files []string) map[string]interface{} {
	var resultValues map[string]interface{}
	for _, file := range files {

		fileExt := strings.ToLower(filepath.Ext(file)[1:])
		log.Infof("Reading configuration file: %s", file)
		log.Debugf("File type: %s", fileExt)

		var override map[string]interface{}
		bs, err := ioutil.ReadFile(file)
		if err != nil {
			log.Errorf("Error processing file: %s, %s", file, err.Error())
			continue
		}

		switch fileExt {
		case "yaml":
			if err := yaml.Unmarshal(bs, &override); err != nil {
				log.Errorf("Error unmarshalling file: %s, %s", file, err.Error())
				continue
			}
		case "yml":
			if err := yaml.Unmarshal(bs, &override); err != nil {
				log.Errorf("Error unmarshalling file: %s, %s", file, err.Error())
				continue
			}
		case "json":
			if err := json.Unmarshal(bs, &override); err != nil {
				log.Errorf("Error unmarshalling file: %s, %s", file, err.Error())
				continue
			}
		default:
			log.Errorf("Unknown file: %s, this cannot be processed", file)
		}
		if err := mergo.Map(&resultValues, override, mergo.WithAppendSlice); err != nil {
			log.Errorf("Error merging file: %s, %s", file, err.Error())
			continue
		}
	}
	return resultValues
}

func validateMetricsQueries(c *config.All) {
	log.Info("Validating Metrics Queries")
	if len(c.MetricDataQueries) > 0 {
		log.Infof("Total metrics queries: %v", len(c.MetricDataQueries))
	} else {
		log.Fatal("Metrics Queries are empty, you need to defined at least one metric in metrics file")
	}
}
