/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// metricsCmd represents the metrics command
var (
	conf       config.All
	metricsCmd = &cobra.Command{
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

	metricsGetCmd.PersistentFlags().StringVar(&conf.Credentials.Profile, "profile", "", "The AWS CLI profile nae from .aws/config or .aws/credential")
	viper.BindPFlag("credentials.profile", metricsGetCmd.PersistentFlags().Lookup("profile"))
}

func get(cmd *cobra.Command, args []string) {
	fmt.Printf("get called with args %v\n", args)
	initConf()
	fmt.Println(conf)
	/*	profile, err := cmd.Flags().GetString("profile")
		if err != nil {
			log.Error(err)
		}

		log.Debugf("AWS Profile: %v", profile)
	*/
	sess, _ := aws.NewSession(&conf.Credentials)
	svc := cloudwatch.New(sess)
	_ = svc
}

func initConf() {
	parseConfFiles(&conf, "metrics")
	parseConfFiles(&conf, "credentials")
	flag.Parse()
	fmt.Println(conf.ToJson())
	fmt.Println(conf.ToYaml())

}

func parseConfFiles(c *config.All, file string) {

	viper.SetConfigName(file)
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Read conf from metrics.yaml file
	err := viper.Unmarshal(&c)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
}
