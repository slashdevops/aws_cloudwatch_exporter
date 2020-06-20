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
	"net/http/pprof"
	"os"
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
		Short: "Useful to debug your metrics defined into the queries files.",
		Long: `The set of commands defined here are usefully to check and test the metrics
defined into your metrics queries files.`,
	}

	metricsGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Execute a call to AWS CloudWatch API to get metrics defined into the metrics queries files.",
		Long: `Using this command you can execute a call to AWS CloudWatch API to get metrics
defined into the metrics queries files.`,
		Run: func(cmd *cobra.Command, args []string) {
			getCmd(cmd, args)
		},
	}

	metricsDisplayPromDescCmd = &cobra.Command{
		Use:   "display",
		Short: "Display prometheus metrics description build from metrics queries files.",
		Long: `Using this command you can display prometheus metrics description build from metrics queries files
into prometheus format, will be the exactly information you will see as a metric name, help string and dimensions.`,
		Run: func(cmd *cobra.Command, args []string) {
			displayPromDescCmd(cmd, args)
		},
	}

	metricsCollectCmd = &cobra.Command{
		Use:   "collect",
		Short: "Start a basic web server with the collector working and every request is sent to AWS CloudWatch API to collect metrics.",
		Long: `Using this command you can start a basic web server with the collector working and every request is sent 
to AWS CloudWatch API to collect metrics,  this work as the command "server start", but you don't 
need to use it as a replacement of the last command, this will be used only to test the result 
before moving to production state.`,
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
	metricsGetCmd.PersistentFlags().StringVar(&conf.Application.MetricStatPeriod, "metricStatPeriod", "5m", "The AWS CloudWatch metrics query stats period")
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

	metricsCollectCmd.Flags().StringP("address", "", "127.0.0.1", "Server address, empty means all addresses")
	metricsCollectCmd.Flags().StringP("port", "", "8080", "Server port")
}

func getCmd(cmd *cobra.Command, args []string) {

	loadFromMetricsFiles(&conf)
	validateMetricsQueries(&conf)

	log.Debugf("Available configuration: %s", conf.ToJSON())
	log.Debugf("Available Env Vars: %s", os.Environ())

	startTime, endTime, period := metrics.GetTimeStamps(time.Now(), conf.Application.MetricStatPeriod, conf.Application.MetricTimeWindow)
	log.Debugf("Start Time: %s", startTime.Format(time.RFC3339))
	log.Debugf("End Time: %s", endTime.Format(time.RFC3339))
	log.Debugf("Period in seconds: %v s", int64(period/time.Second))

	m := metrics.New(&conf)
	mdi := m.GetMetricDataInput(startTime, endTime, period, "")

	log.Debugf("Metrics queries: %s", mdi.String())

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

	log.Debugf("Available configuration: %s", conf.ToJSON())
	log.Debugf("Available Env Vars: %s", os.Environ())

	m := metrics.New(&conf)

	for _, md := range m.GetMetricsDesc() {
		log.Println(md.String())
	}
}

func collectCmd(cmd *cobra.Command, args []string) {

	loadFromMetricsFiles(&conf)
	validateMetricsQueries(&conf)

	log.Debugf("Available configuration: %s", conf.ToJSON())
	log.Debugf("Available Env Vars: %s", os.Environ())

	m := metrics.New(&conf)
	sess := awshelper.NewSession(&conf.AWS)
	cwc := cloudwatch.New(sess)

	c := collector.New(&conf, m, cwc)

	prometheus.MustRegister(c)
	mux := http.NewServeMux()

	// metrics path
	mux.Handle("/metrics", promhttp.Handler())

	// Debug & Profiling
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	mux.HandleFunc("/debug/pprof/mutex", pprof.Index)
	mux.HandleFunc("/debug/pprof/goroutine", pprof.Index)
	mux.HandleFunc("/debug/pprof/threadcreate", pprof.Index)
	mux.HandleFunc("/debug/pprof/block", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// default root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
            <head><title>` + appName + ` Exporter</title></head>
            <body>
            <h1>` + appName + ` Exporter</h1>
            <p><a href='/metrics'>Metrics</a></p>
			<h2>Debug and profile</h2>
            <p><a href='/debug/pprof/'>/debug/pprof</a></p>
            <p><a href='/debug/pprof/heap'>/debug/pprof/heap</a></p>
            <p><a href='/debug/pprof/mutex'>/debug/pprof/mutex</a></p>
            <p><a href='/debug/pprof/goroutine'>/debug/pprof/goroutine</a></p>
            <p><a href='/debug/pprof/threadcreate'>/debug/pprof/threadcreate</a></p>
            <p><a href='/debug/pprof/block'>/debug/pprof/block</a></p>
            <p><a href='/debug/pprof/cmdline'>/debug/pprof/cmdline</a></p>
            <p><a href='/debug/pprof/profile'>/debug/pprof/profile</a></p>
            <p><a href='/debug/pprof/symbol'>/debug/pprof/symbol</a></p>
            <p><a href='/debug/pprof/trace'>/debug/pprof/trace</a></p>
            </body>
            </html>`))
		if err != nil {
			log.Error(err)
		}
	})

	a, _ := cmd.Flags().GetString("address")
	p, _ := cmd.Flags().GetString("port")
	soc := fmt.Sprintf("%s:%v", a, p)

	log.Infof("Starting test server on %s", soc)
	log.Warn("Don't use this server as your default server used for prometheus exporter, instead use 'server start' command.")
	log.Fatal(http.ListenAndServe(soc, mux))
}
