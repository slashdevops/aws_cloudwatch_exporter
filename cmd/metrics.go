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
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/awshelper"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/metrics"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// metricsCmd represents the metrics command
var (
	conf           config.All
	serverConfFile string = "server.yaml"
	metricsCmd            = &cobra.Command{
		Use:   "metrics [COMMANDS]",
		Short: "useful to get metrics",
		Long:  `metrics commands`,
	}

	metricsGetCmd = &cobra.Command{
		Use:   "get [OPTIONS] [ARGS]",
		Short: "get metrics",
		Long:  `Get metrics from CloudWatch using the metrics queries defined in the yaml files `,
		Run: func(cmd *cobra.Command, args []string) {
			get(cmd, args)
		},
	}
)

func init() {
	rootCmd.AddCommand(metricsCmd)
	metricsCmd.AddCommand(metricsGetCmd)

	// Files
	metricsGetCmd.PersistentFlags().StringVar(&conf.Application.CredentialsFile, "credentialsFile", "credentials.yaml", "The metrics files with the CloudWatch Queries")
	viper.BindPFlag("application.credentialsFile", metricsGetCmd.PersistentFlags().Lookup("credentialsFile"))

	metricsGetCmd.PersistentFlags().StringSliceVar(&conf.Application.MetricsFiles, "metricsFiles", nil, "Metrics files, example: --metricsFile ~/tmp/queries/m1.yaml --metricsFile ~/tmp/queries/m2.yml")
	viper.BindPFlag("application.metricsFiles", metricsGetCmd.PersistentFlags().Lookup("metricsFiles"))

	// Behavior parameters
	metricsGetCmd.PersistentFlags().StringVar(&conf.AWS.Profile, "profile", "", "The AWS CLI profile nae from .aws/config or .aws/credential")
	viper.BindPFlag("aws.profile", metricsGetCmd.PersistentFlags().Lookup("profile"))

	metricsGetCmd.PersistentFlags().StringVar(&conf.Application.StatsPeriod, "statsPeriod", "1m", "The AWS Cloudwatch metrics query stats period")
	viper.BindPFlag("application.statsPeriod", metricsGetCmd.PersistentFlags().Lookup("statsPeriod"))

}

func get(cmd *cobra.Command, args []string) {
	initConf()

	startTime, endTime, period := metrics.GetTimeStamps(time.Now(), conf.Application.StatsPeriod)

	log.Debugf("Start Time: %s", startTime.Format(time.RFC3339))
	log.Debugf("End Time: %s", endTime.Format(time.RFC3339))
	log.Debugf("Period in seconds: %v s", int64(period/time.Second))

	mdi := metrics.NewGetMetricDataInput(&conf.MetricsQueriesConf, startTime, endTime, period, "")
	sess, _ := awshelper.NewSession(&conf.AWS)
	svc := cloudwatch.New(sess)
	mdo, err := svc.GetMetricData(mdi)
	if err != nil {
		log.Errorf("Error getting metrics %v", err)
	}
	fmt.Println(mdo)
}

func initConf() {
	parseConfFiles(&conf)
	parseMetricsFiles(&conf)
	//log.Debugf("Configuration %s", conf.ToJson())
	fmt.Println(conf.ToJson())
	//fmt.Println(conf.ToYaml())
}

// Unmarshall Yaml files into c config structure
func parseConfFiles(c *config.All) {
	viper.SetDefault("application.ServerFile", serverConfFile)

	files := []string{
		serverConfFile,
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
		//viper.SetEnvPrefix("ace")

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
			viper.MergeInConfig()
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
