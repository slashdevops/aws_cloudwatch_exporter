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
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/awshelper"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/collector"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/metrics"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// metricsCmd represents the metrics command
var (
	metricsCmd = &cobra.Command{
		Use:   "metrics [commands]",
		Short: "Useful to debug your metrics",
		Long:  `metrics commands`,
	}

	metricsGetCmd = &cobra.Command{
		Use:   "get [options] [args]",
		Short: "get metrics",
		Long:  `Get metrics from AWS CloudWatch using the metrics queries defined in the [yaml|json] files`,
		Run: func(cmd *cobra.Command, args []string) {
			getCmd(cmd, args)
		},
	}

	metricsDisplayPromDescCmd = &cobra.Command{
		Use:   "display [options] [args]",
		Short: "display promDesc",
		Long:  `Display prometheus metrics definitions`,
		Run: func(cmd *cobra.Command, args []string) {
			displayPromDescCmd(cmd, args)
		},
	}

	metricsCollectCmd = &cobra.Command{
		Use:   "collect [options] [args]",
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

	// local flags
	metricsGetCmd.Flags().StringP("outFormat", "", "yaml", "Output format for results, possible values: [yaml|json]")
	metricsGetCmd.Flags().StringP("outFile", "", "", "Output file where to store the results.")

	metricsCollectCmd.Flags().StringP("address", "", "127.0.0.1", "Test server address, empty means all addresses")
	metricsCollectCmd.Flags().StringP("port", "", "8080", "Test server port")
}

func getCmd(cmd *cobra.Command, args []string) {

	loadFromMetricsFiles(&conf)
	validateMetricsQueries(&conf)

	if conf.Server.Debug {
		log.Debug(conf.ToJSON())
	}

	startTime, endTime, period := metrics.GetTimeStamps(time.Now(), conf.Application.MetricStatPeriod, conf.Application.MetricTimeWindow)
	log.Debugf("Start Time: %s", startTime.Format(time.RFC3339))
	log.Debugf("End Time: %s", endTime.Format(time.RFC3339))
	log.Debugf("Period in seconds: %v s", int64(period/time.Second))

	m := metrics.New(&conf)
	mdi := m.GetMetricDataInput(startTime, endTime, period, "")

	sess := awshelper.NewSession(&conf.AWS)
	svc := cloudwatch.New(sess)
	mdo, err := svc.GetMetricData(mdi)
	if err != nil {
		log.Fatalf("Error getting metrics: %v", err)
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

	outFile, _ := cmd.Flags().GetString("outFile")
	if outFile != "" {
		if err := ioutil.WriteFile(outFile, outMetrics, 0644); err != nil {
			log.Panic(err)
		}
	} else {
		log.Println(string(outMetrics))
	}
}

func displayPromDescCmd(cmd *cobra.Command, args []string) {

	loadFromMetricsFiles(&conf)
	validateMetricsQueries(&conf)

	if conf.Server.Debug {
		log.Debug(conf.ToJSON())
	}

	m := metrics.New(&conf)

	for _, md := range m.GetMetricsDesc() {
		log.Println(md.String())
	}
}

func collectCmd(cmd *cobra.Command, args []string) {

	loadFromMetricsFiles(&conf)
	validateMetricsQueries(&conf)

	if conf.Server.Debug {
		log.Debug(conf.ToJSON())
	}

	m := metrics.New(&conf)
	sess := awshelper.NewSession(&conf.AWS)
	cwc := cloudwatch.New(sess)

	c := collector.New(&conf, m, cwc)

	prometheus.MustRegister(c)
	http.Handle("/metrics", promhttp.Handler())

	a, _ := cmd.Flags().GetString("address")
	p, _ := cmd.Flags().GetString("port")
	soc := fmt.Sprintf("%s:%v", a, p)
	log.Infof("Starting test server on %s", soc)
	log.Fatal(http.ListenAndServe(soc, nil))
}
