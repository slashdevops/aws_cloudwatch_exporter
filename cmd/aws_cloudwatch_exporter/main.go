package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/server"
	"github.com/slashdevops/aws_cloudwatch_exporter/web"

	flag "github.com/spf13/pflag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/spf13/viper"
)

const (
	namespace                = "aws_cloudwatch"
	appName                  = "aws_cloudwatch_exporter"
	appDescription           = "AWS CloudWatch Exporter for Prometheus metrics"
	appMetricsPath           = "/metrics"
	appConfigServerFileName  = "server"
	appConfigMetricsFileName = "metrics"
	appConfigCredFileName    = "credentials"
)

var (
	showVersion = flag.Bool("version", false, "Print version information.")
	serverAddr  = flag.String("server.address", ":", "Address to listen on for web interface and telemetry.")
	serverPort  = flag.String("server.port", "9690", "Port to listen on for web interface and telemetry.")
)

func init() {
	prometheus.MustRegister(version.NewCollector(namespace))
}

func main() {
	var conf config.All

	logger := logrus.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.DebugLevel)

	vServer := viper.New()
	vApp := viper.New()
	vMetrics := viper.New()
	vCreds := viper.New()

	vApp.SetDefault("application.name", appName)
	vApp.SetDefault("application.description", appDescription)
	vApp.SetDefault("application.logger", logger)

	// Read conf from server.yaml file
	err := vApp.Unmarshal(&conf)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	vServer.SetDefault("server.metricsPath", appMetricsPath)
	vServer.SetConfigName(appConfigServerFileName)
	vServer.AddConfigPath(".")
	vServer.AddConfigPath("/etc/" + appName)
	vServer.AddConfigPath("$HOME/." + appName)
	vServer.AutomaticEnv()
	vServer.SetConfigType("yaml")
	vServer.SetConfigType("yml")

	if err := vServer.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Read conf from server.yaml file
	err = vServer.Unmarshal(&conf)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	vMetrics.SetConfigName(appConfigMetricsFileName)
	vMetrics.AddConfigPath(".")
	vMetrics.AddConfigPath("/etc/" + appName)
	vMetrics.AddConfigPath("$HOME/." + appName)
	vMetrics.AutomaticEnv()
	vMetrics.SetConfigType("yaml")
	vMetrics.SetConfigType("yml")

	if err := vMetrics.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Read conf from metrics.yaml file
	err = vMetrics.Unmarshal(&conf)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	vCreds.SetConfigName(appConfigCredFileName)
	vCreds.AddConfigPath(".")
	vCreds.AddConfigPath("/etc/" + appName)
	vCreds.AddConfigPath("$HOME/." + appName)
	vCreds.AutomaticEnv()
	vCreds.SetConfigType("yaml")
	vCreds.SetConfigType("yml")

	if err := vCreds.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	// Read conf from metrics.yaml file
	err = vCreds.Unmarshal(&conf)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version.Print(namespace))
		os.Exit(0)
	}

	logger.Debug("conf: %v", conf)

	h := web.NewHandlers(&conf)
	mux := http.NewServeMux()
	h.SetupRoutes(mux)
	s := server.New(mux, &conf)

	ctx := context.Background()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	logger.Printf("Starting %s %s %s on %s", appDescription, version.Info(), version.BuildContext(), *serverPort)

	// Start the server
	go func() {
		if err := s.ListenAndServe(); err != nil {
			logger.Printf("Error starting %s", appDescription)
			os.Exit(1)
		}
	}()

	// create a context for graceful http server shutdown
	srvCtx, srvCancel := context.WithTimeout(ctx, 5*time.Second)
	defer srvCancel()
	<-c

	logger.Printf("Shutting Down %s signal received", appDescription)
	err = s.Shutdown(srvCtx)
	if err != nil {
		logger.Fatalf("Error shuting down %s", appDescription)
	}
}
