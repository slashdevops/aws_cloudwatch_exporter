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
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slashdevops/aws_cloudwatch_exporter/collector"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/awshelper"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/metrics"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Server commands",
		Long:  `A longer description that spans `,
	}

	serverStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the http server",
		Long:  `This start the http server to handle connections for metrics endpoint`,
		Run: func(cmd *cobra.Command, args []string) {
			startCmd(cmd, args)
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd)
}

func startCmd(cmd *cobra.Command, args []string) {

	m := metrics.New(&conf)
	sess, _ := awshelper.NewSession(&conf.AWS)
	c := collector.New(&conf, m, sess)
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())

	log.Info("Starting Server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
