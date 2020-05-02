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

	"github.com/aimroot/aws_cloudwatch_exporter/internal/server"
	_ "github.com/aimroot/aws_cloudwatch_exporter/internal/server"
	"github.com/aimroot/aws_cloudwatch_exporter/web"

	flag "github.com/spf13/pflag"

	"github.com/aimroot/aws_cloudwatch_exporter/config"
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
	logger := logrus.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.DebugLevel)

	vServer := viper.New()
	vMetrics := viper.New()

	// Default Variables
	vServer.SetDefault("app.name", appName)
	vServer.SetDefault("app.description", appDescription)
	vServer.SetDefault("server.metricsPath", appMetricsPath)

	var sConf config.Server
	sConf.Logger = logger

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

	err := vServer.Unmarshal(&sConf)
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

	var mConf config.Metrics

	if err := vMetrics.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	err = vMetrics.Unmarshal(&mConf)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	flag.Parse()

	listenAddr := *serverAddr + *serverPort

	if *showVersion {
		fmt.Println(version.Print(namespace))
		os.Exit(0)
	}

	logger.Debug("sConf: %v", sConf)
	logger.Debug("mConf: %v", mConf)

	h := web.NewHandlers(&sConf)
	mux := http.NewServeMux()
	h.SetupRoutes(mux)
	server := server.New(mux, listenAddr)

	ctx := context.Background()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	logger.Printf("Starting %s %s %s on %s", appDescription, version.Info(), version.BuildContext(), *serverPort)

	// Start the server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Printf("Error starting %s", appDescription)
			os.Exit(1)
		}
	}()

	// create a context for graceful http server shutdown
	srvCtx, srvCancel := context.WithTimeout(ctx, 5*time.Second)
	defer srvCancel()
	<-c

	logger.Printf("Shutting Down %s signal received", appDescription)
	err = server.Shutdown(srvCtx)
	if err != nil {
		logger.Fatalf("Error shuting down %s", appDescription)
	}
}
