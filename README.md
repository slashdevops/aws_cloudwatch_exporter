# aws_cloudwatch_exporter

Prometheus exporter for AWS CloudWatch

## Config files

* server.yaml
* credentials.yaml
* metrics.yaml

## Building and release tools

* https://github.com/prometheus/promu

## References

* https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/

## Development Notes

### Metrics suffix

AWS CloudWatch GetMetricsData Stat conversion to prometheus suffix
Minimum     -->
Maximum     -->
Sum         --> _sum
Average     --> 
SampleCount --> _count

### Exporter metrics

https://prometheus.io/docs/practices/naming/

```text
<metric name>{<label name>=<label value>, ...}

<exporter>_build_info
<exporter>_scrape_duration_seconds
<exporter>_requests_total {state="successful|failed", ...}
<exporter>_request_duration_seconds {state="successful|failed", ...}
<exporter>_points_total {state="successful|failed", ...}
<exporter>_metrics_total {state="successful|failed", ...}
```