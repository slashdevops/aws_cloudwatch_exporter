/*
Copyright © 2020 NAME HERE christian@slashdevops.com

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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Namespace      = "aws_cloudwatch_exporter"
	appName        = "aws_cloudwatch_exporter"
	appDescription = "AWS CloudWatch Exporter for Gatherer metrics"
	appMetricsPath = "/metrics"
)

// rootCmd represents the base command when called without any subcommands
var (
	// Global vars
	conf config.All

	rootCmd = &cobra.Command{
		Use:   "aws_cloudwatch_exporter",
		Short: "AWS CloudWatch Exporter for prometheus.io",
		Long:  `aws_cloudwatch_exporter is an AWS CloudWatch exporter for prometheus.io`,
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

	// Files
	rootCmd.PersistentFlags().StringVar(&conf.ServerFile, "serverFile", "server.yaml", "The server configuration file")
	if err := viper.BindPFlag("application.serverFile", rootCmd.PersistentFlags().Lookup("serverFile")); err != nil {
		log.Error(err)
	}

	rootCmd.PersistentFlags().StringVar(&conf.Application.CredentialsFile, "credentialsFile", "credentials.yaml", "The metrics files with the AWS CloudWatch Metrics Queries")
	if err := viper.BindPFlag("application.credentialsFile", rootCmd.PersistentFlags().Lookup("credentialsFile")); err != nil {
		log.Error(err)
	}

	rootCmd.PersistentFlags().StringSliceVar(&conf.Application.MetricsFiles, "metricsFiles", nil, "Metrics files, example: --metricsFile ~/tmp/queries/m1.yaml --metricsFile ~/tmp/queries/m2.yml")
	if err := viper.BindPFlag("application.metricsFiles", rootCmd.PersistentFlags().Lookup("metricsFiles")); err != nil {
		log.Error(err)
	}

	//
	rootCmd.PersistentFlags().StringVar(&conf.AWS.Profile, "profile", "", "The AWS CLI profile name defined in .aws/config or .aws/credential")
	if err := viper.BindPFlag("aws.profile", rootCmd.PersistentFlags().Lookup("profile")); err != nil {
		log.Error(err)
	}

}

func initConfig() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)
	// Set the output of the message for the current logrus instance,
	// Output of logrus instance can be set to any io.writer
	log.Out = os.Stdout

	parseConfFiles(&conf)
	parseMetricsFiles(&conf)
	// fmt.Println(conf.ToJson())
	fmt.Println(conf.ToJson())
	// fmt.Println(conf.ToYaml())
}

// Unmarshall Yaml files into c config structure
func parseConfFiles(c *config.All) {
	conf.Application.Name = appName
	conf.Application.Namespace = Namespace
	conf.Application.Description = appDescription
	conf.Application.MetricsPath = appMetricsPath
	conf.Application.Version = version.Version
	conf.Application.Revision = version.Revision
	conf.Application.GoVersion = version.GoVersion
	conf.Application.BuildUser = version.BuildUser
	conf.Application.BuildDate = version.BuildDate
	conf.Application.Info = version.Info()
	conf.Application.BuildInfo = version.BuildContext()

	// Config files to be load
	files := []string{
		conf.ServerFile,
		c.Application.CredentialsFile,
	}

	for _, file := range files {
		log.Debugf("Configuration file: %s", file)
		log.Debugf("file: %s", filepath.Base(file))
		log.Debugf("Location: %s", filepath.Dir(file))
		log.Debugf("Kind: %s", filepath.Ext(file)[1:])

		viper.SetConfigName(filepath.Base(file))
		viper.AddConfigPath(filepath.Dir(file))
		viper.SetConfigType(filepath.Ext(file)[1:])

		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		log.Debugf("Reading configuration from file: %s", file)
		if err := viper.ReadInConfig(); err != nil {
			log.Errorf("Error reading config file, %s", err)
		}

		log.Debug("Filling conf structure")
		err := viper.Unmarshal(&c)
		if err != nil {
			log.Errorf("Unable to decode into struct, %v", err)
		}
	}
}

// Unmarshall Yaml files into c config structure
func parseMetricsFiles(c *config.All) {

	for i, file := range c.Application.MetricsFiles {
		log.Debugf("Configuration file: %s", file)
		log.Debugf("file: %s", filepath.Base(file))
		log.Debugf("Location: %s", filepath.Dir(file))
		log.Debugf("Kind: %s", filepath.Ext(file)[1:])

		viper.SetConfigName(filepath.Base(file))
		viper.AddConfigPath(filepath.Dir(file))
		viper.SetConfigType(filepath.Ext(file)[1:])

		if i < 1 {
			log.Debugf("Reading configuration from file: %s", file)
			if err := viper.ReadInConfig(); err != nil {
				log.Errorf("Error reading config file, %s", err)
			}
		} else {
			log.Debugf("Merging configuration of file: %s", file)
			if err := viper.MergeInConfig(); err != nil {
				log.Errorf("Error merging config file, %s", err)
			}
		}
	}

	if len(c.Application.MetricsFiles) > 0 {
		log.Debug("Filling conf structure")
		err := viper.Unmarshal(&c)
		if err != nil {
			log.Errorf("Unable to decode into struct, %v", err)
		}
	} else {
		log.Errorf("Metrics configuration file: \"%v\" doesn't exist", c.Application.MetricsFiles)
	}
}
