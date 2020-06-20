# server file configuration

This is the server configuration file options

```yaml
server:                               # This is related to the http service configuration
  address: "127.0.0.1"                # Type: string, IP Address in the host where you want the service listen 
  port: 9690                          # Type: int, Port in the host where you want the service listen
  readTimeout: 2s                     # Type: time.Duration, ReadTimeout is the maximum duration for reading the entire request, including the body. see: https://golang.org/pkg/net/http/
  writeTimeout: 5s                    # Type: time.Duration, WriteTimeout is the maximum duration before timing out writes of the response. see: https://golang.org/pkg/net/http/ 
  idleTimeout: 60s                    # Type: time.Duration, IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled. see: https://golang.org/pkg/net/http/
  ReadHeaderTimeout: 5s               # Type: time.Duration, ReadHeaderTimeout is the amount of time allowed to read request headers. see: https://golang.org/pkg/net/http/
  shutdownTimeout: 30s                # Type: time.Duration, The time you want to wait until connections established finish before shutdown the server
  KeepAlivesEnabled: true             # Type: boolean, KeepAlivesEnabled controls whether HTTP keep-alives are enabled. see: https://golang.org/pkg/net/http/
  LogFormat: text                     # Type: string, Define the log output format of the server, valid values [text|json]
  Debug: false                        # Type: boolean, If this is enabled, the log debug messages are visible in the log output

application:                          # This is related to the application behavior
  metricStatPeriod: 5m                # Type: time.Duration, Defined the global period of time .see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricStat.html
  metricTimeWindow: 10m               # Type: time.Duration, Defined the time windows between the StartTime and EndTime. see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
  metricsFiles:                       # Type: Array, List of files with the definitions of metrics queries 
    - metrics.yaml                    # Type: string, Part of the array list with the location/path of file with the metrics queries in the format defined in metrics.md file
```

## Help links

for **readTimeout, writeTimeout, idleTimeout, ReadHeaderTimeout and KeepAlivesEnabled**

* https://golang.org/pkg/net/http/

for **metricStatPeriod and metricTimeWindow**

* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricStat.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_Range.html
* https://docs.aws.amazon.com/cli/latest/reference/cloudwatch/get-metric-data.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html

for **metricsFiles**

* [metrics.md](metrics.md)
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html