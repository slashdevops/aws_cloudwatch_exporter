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
	"gopkg.in/yaml.v3"
)

const (
	appName        = "aws_cloudwatch_exporter"
	appDescription = `AWS CloudWatch exporter for prometheus.io
This exporter use GetMetricData API to get the metrics from AWS CloudWatch`
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
	rootCmd.PersistentFlags().BoolVar(&conf.Server.Debug, "debug", false, "If this is enabled, the log debug messages are visible in the log output")
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

	rootCmd.PersistentFlags().StringSliceVar(&conf.Application.MetricsFiles, "metricsFiles", []string{"metrics.yaml"}, "Metrics files, example: --metricsFiles ~/tmp/queries/m1.yaml --metricsFiles ~/tmp/queries/m2.yml")
	if err := viper.BindPFlag("application.metricsFiles", rootCmd.PersistentFlags().Lookup("metricsFiles")); err != nil {
		log.Error(err)
	}

	// AWS Credentials conf
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
}

func initConfig() {
	if strings.ToLower(conf.Server.LogFormat) == "json" {
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

// Unmarshall Yaml files into c config structure
func loadFromConfigFiles(fileName string, c *config.All) {

	if !fileExists(fileName) {
		log.Warnf("The file %s doesn't exist, I will try to use configuration values from flags or ENV vars", fileName)
	}

	log.Infof("Reading configuration file: %s", fileName)

	// fileNameNoExt := strings.TrimSuffix(file, filepath.Ext(file))

	log.Debugf("Parsing configuration file path: %s", fileName)
	log.Debugf("File: %s", filepath.Base(fileName))
	// log.Debugf("File without ext: %s", fileNameNoExt)
	log.Debugf("Location: %s", filepath.Dir(fileName))
	log.Debugf("File ext: %s", filepath.Ext(fileName)[1:])

	// viper.SetConfigName(fileNameNoExt)
	viper.SetConfigName(filepath.Base(fileName))
	viper.AddConfigPath(filepath.Dir(fileName))
	viper.SetConfigType(filepath.Ext(fileName)[1:])

	// Read env vars equals as the mapstructure defined into the config.go
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	log.Debugf("Loading configuration from file: %s", fileName)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	log.Debugf("Filling configuration structure from file: %s", fileName)
	err := viper.Unmarshal(&c)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

}

// Unmarshall Yaml files into c config structure
// NOTE: Unfortunately viper.MergeInConfig() does the merge using override, so
// this is the reason to do not user it.
func loadFromMetricsFiles(c *config.All) {

	metricsQueries := MergeMetricsFiles(c.Application.MetricsFiles)

	if err := viper.MergeConfigMap(metricsQueries); err != nil {
		log.Fatalf("Error merging MetricsQueries read from files into config structure, check your metrics queries: %s", err.Error())
	}

	if len(c.Application.MetricsFiles) > 0 {
		log.Debugf("Filling configuration structure from metrics queries file: %s", c.Application.MetricsFiles)
		err := viper.Unmarshal(&c)
		if err != nil {
			log.Fatalf("Unable to unmarshal Metrics queries files into config struct, %s", err.Error())
		}
	} else {
		log.Fatal("Metrics queries files don't provided, you need to provide at least one to continue")
	}
}

// This function merge files with metrics queries into a map without override keys
func MergeMetricsFiles(files []string) map[string]interface{} {
	var resultValues map[string]interface{}
	for _, file := range files {

		if !fileExists(file) {
			log.Fatalf("The file %s does not exist, you need to provide valid metrics queries file", file)
		}

		fileExt := strings.ToLower(filepath.Ext(file)[1:])
		log.Infof("Reading metrics queries file: %s", file)
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
