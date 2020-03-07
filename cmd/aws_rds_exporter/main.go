package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"text/template"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/aimroot/aws_cloudwatch_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/spf13/viper"
)

const (
	namesapce      = "aws_cloudwatch"
	appName        = "aws_cloudwatch_exporter"
	appDescription = "AWS CloudWatch Exporter"
)

var (
	showVersion = flag.Bool("version", false, "Print version information.")
	serverAddr  = flag.String("server.address", ":", "Address to listen on for web interface and telemetry.")
	serverPort  = flag.String("server.port", "9690", "Address to listen on for web interface and telemetry.")
	//role_arn    = flag.String("role_arn", "", "AWS Role ARN.  ENV VAR ROLE_ARN")
	metricsPath = flag.String("server.metricsPath", "/metrics", "Path under which to expose metrics.")
	configFile  = flag.String("server.configFile", "config.yaml", "This exporter configuration file name.")
)

func init() {
	prometheus.MustRegister(version.NewCollector(namesapce))
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/" + appName)
	viper.AddConfigPath("$HOME/." + appName)
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetConfigType("yml")

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
		fmt.Println(version.Print(namesapce))
		os.Exit(0)
	}

	mux := http.NewServeMux()
	server := &http.Server{
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Addr:              listenAddr,
		Handler:           mux,
	}

	ctx := context.Background()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/healthz", healthHandler)
	mux.Handle(*metricsPath, promhttp.Handler())

	// Debug & Profiling
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	log.Printf("Starting %s %s %s on %s", appDescription, version.Info(), version.BuildContext(), *serverPort)

	// Start the server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Printf("Error starting %s", appDescription)
			os.Exit(1)
		}
	}()

	// create a context for graceful http server shutdown
	srvCtx, srvCancel := context.WithTimeout(ctx, 5*time.Second)
	defer srvCancel()
	<-c
	log.Printf("Shutting Down %s signal received", appDescription)
	err = server.Shutdown(srvCtx)
	if err != nil {
		log.Fatalf("Error shuting down %s", appDescription)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title         string
		MetricHandler string
	}{appDescription, *metricsPath}
	t := template.Must(template.ParseFiles("../../web/template/index.html"))
	t.Execute(w, data)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}
