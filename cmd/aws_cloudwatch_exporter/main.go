package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

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
	namespace         = "aws_cloudwatch"
	appName           = "aws_cloudwatch_exporter"
	appDescription    = "AWS CloudWatch Exporter for Prometheus metrics"
	appConfigFileName = "config"
)

var (
	showVersion = flag.Bool("version", false, "Print version information.")
	serverAddr  = flag.String("server.address", ":", "Address to listen on for web interface and telemetry.")
	serverPort  = flag.String("server.port", "9690", "Port to listen on for web interface and telemetry.")
	//role_arn    = flag.String("role_arn", "", "AWS Role ARN.  ENV VAR ROLE_ARN")
	configFile = flag.String("server.configFile", "config.yaml", "This exporter configuration file name.")
)

func init() {
	prometheus.MustRegister(version.NewCollector(namespace))
}

func main() {
	logger := log.New(os.Stdout, namespace, log.LstdFlags|log.Lshortfile)

	viper.SetConfigName(appConfigFileName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/" + appName)
	viper.AddConfigPath("$HOME/." + appName)
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigType("yml")

	viper.SetDefault("app.name", appName)
	viper.SetDefault("app.description", appDescription)

	var conf config.Config

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&conf)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
	flag.Parse()

	listenAddr := *serverAddr + *serverPort

	if *showVersion {
		fmt.Println(version.Print(namespace))
		os.Exit(0)
	}

	h := web.NewHandlers(logger, &conf)
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
