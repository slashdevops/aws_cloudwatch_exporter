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
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slashdevops/aws_cloudwatch_exporter/collector"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/awshelper"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/metrics"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// metricsCmd represents the metrics command
var (
	metricsCmd = &cobra.Command{
		Use:   "metrics [COMMANDS]",
		Short: "useful to get metrics",
		Long:  `metrics commands`,
	}

	metricsGetCmd = &cobra.Command{
		Use:   "get [OPTIONS] [ARGS]",
		Short: "get metrics",
		Long:  `Get metrics from CloudWatch using the metrics queries defined in the yaml files`,
		Run: func(cmd *cobra.Command, args []string) {
			getCmd(cmd, args)
		},
	}

	metricsDisplayPromDescCmd = &cobra.Command{
		Use:   "display [OPTIONS] [ARGS]",
		Short: "display promDesc",
		Long:  `Display prometheus metrics definitions`,
		Run: func(cmd *cobra.Command, args []string) {
			displayPromDescCmd(cmd, args)
		},
	}

	metricsCollectCmd = &cobra.Command{
		Use:   "collect [OPTIONS] [ARGS]",
		Short: "collect metrics",
		Long:  `Collect metrics from AWS CloudWatch using prometheus collector`,
		Run: func(cmd *cobra.Command, args []string) {
			collectCmd(cmd, args)
		},
	}
)

func init() {
	rootCmd.AddCommand(metricsCmd)
	metricsCmd.AddCommand(metricsGetCmd)
	metricsCmd.AddCommand(metricsDisplayPromDescCmd)
	metricsCmd.AddCommand(metricsCollectCmd)

	// Behavior parameters
	metricsGetCmd.PersistentFlags().StringVar(&conf.Application.MetricStatPeriod, "metricStatPeriod", "5m", "The AWS Cloudwatch metrics query stats period")
	if err := viper.BindPFlag("application.metricStatPeriod", metricsGetCmd.PersistentFlags().Lookup("metricStatPeriod")); err != nil {
		log.Error(err)
	}
	metricsGetCmd.PersistentFlags().StringVar(&conf.Application.MetricTimeWindow, "metricTimeWindow", "10m", "Time windows gap used to get metrics stats")
	if err := viper.BindPFlag("application.metricTimeWindow", metricsGetCmd.PersistentFlags().Lookup("metricTimeWindow")); err != nil {
		log.Error(err)
	}

	// Output parameters
	metricsGetCmd.Flags().StringP("outFormat", "", "yaml", "Output format for results. (supported [yaml|json] only)")
	metricsGetCmd.Flags().StringP("outFile", "", "", "Filename Save the result")
}

func getCmd(cmd *cobra.Command, args []string) {
	startTime, endTime, period := metrics.GetTimeStamps(time.Now(), conf.Application.MetricStatPeriod, conf.Application.MetricTimeWindow)
	log.Debugf("Start Time: %s", startTime.Format(time.RFC3339))
	log.Debugf("End Time: %s", endTime.Format(time.RFC3339))
	log.Debugf("Period in seconds: %v s", int64(period/time.Second))

	m := metrics.New(&conf)
	mdi := m.GetMetricDataInput(startTime, endTime, period, "")

	sess, _ := awshelper.NewSession(&conf.AWS)
	svc := cloudwatch.New(sess)
	mdo, err := svc.GetMetricData(mdi)
	if err != nil {
		log.Errorf("Error getting metrics %v", err)
	}

	var outMetrics []byte

	outFormat, _ := cmd.Flags().GetString("outFormat")
	if outFormat == "yaml" {
		out, err := yaml.Marshal(mdo)
		if err != nil {
			log.Panicln(err)
		}
		outMetrics = out
	} else if outFormat == "json" {
		out, err := json.MarshalIndent(mdo, "", " ")
		if err != nil {
			log.Panicln(err)
		}
		outMetrics = out
	} else {
		log.Errorf("Invalid flag value outFormat: %s", outFormat)
	}

	outFile, _ := cmd.Flags().GetString("outFormat")
	if outFile != "" {
		if err := ioutil.WriteFile(outFile, outMetrics, 0644); err != nil {
			log.Panic(err)
		}
	} else {
		fmt.Println(string(outMetrics))
	}
}

func displayPromDescCmd(cmd *cobra.Command, args []string) {
	m := metrics.New(&conf)

	for _, md := range m.GetMetricsDesc() {
		fmt.Println(md.String())
	}
}

func collectCmd(cmd *cobra.Command, args []string) {
	m := metrics.New(&conf)
	sess, _ := awshelper.NewSession(&conf.AWS)

	c := collector.New(&conf, m, sess)
	prometheus.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())

	log.Info("Starting Server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
