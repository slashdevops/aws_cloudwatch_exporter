# aws_cloudwatch_exporter

![Main branch workflow](https://github.com/slashdevops/aws_cloudwatch_exporter/workflows/main%20branch%20workflow/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/slashdevops/aws_cloudwatch_exporter)](https://goreportcard.com/report/github.com/slashdevops/aws_cloudwatch_exporter)
![Docker Pulls](https://img.shields.io/docker/pulls/slashdevops/aws-cloudwatch-exporter)

Prometheus exporter for AWS CloudWatch

This exporter used the [AWS CloudWatch GetMetricsData API call](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html), please it is important you read

* [cloudwatch-getmetricdata-api](https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/)
* [Amazon CloudWatch pricing](https://aws.amazon.com/cloudwatch/pricing/?nc1=h_ls)

**WARNING**:

* The use of this exporter could be expensive, this need be used taking care.

**NOTES**:

* When you use [AWS CloudWatch GetMetricsData API call](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html) with 1 request you can get 500 metrics a time, so
read the links above before use it.

LIMITATIONS:

* You could define 500 metrics queries in total (a future release will allow more)

## Configuration

### Server

To configure the server you have 3 ways:

1. Configuration Files   (i.e.: ./server.yaml)
2. Environment Variables (i.e.: SERVER_PORT, SERVER_ADDRESS, etc)
3. Program Flags         (i.e.: --serverPort, --serverAddress,etc)

### Metrics Queries

This exporter used the standard metrics queries format used by [AWS CloudWatch API GetMetricData](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html)
More information at [metrics.md](docs/metrics.md)

### Docs

* [server.yaml](docs/server.md)
* [metrics.yaml](docs/metrics.md)

### Running

#### Docker

Using your own `~/.aws` configuration in local

```bash
docker run --rm \
    -i \
    -u nobody:nogroup \
    --privileged \
    -v ~/tmp/queries/m1.yaml:/home/nobody/m1.yaml:ro \
    -v ~/tmp/queries/m2.yaml:/home/nobody/m2.yaml:ro \
    -v /tmp/:/home/nobody/tmp/:rw \
    -v ~/.aws:/home/nobody/.aws:ro \
    -e "AWS_SDK_LOAD_CONFIG=true" \
    -e "AWS_PROFILE=slashdevops" \
    slashdevops/aws-cloudwatch-exporter-linux-amd64:develop metrics get --metricsFiles /home/nobody/m1.yaml --outFile /home/nobody/tmp/out.yaml
```

Check the result

```bash
cat /tmp/out.yaml
```

#### Binary

If you download the binary from releases

```bash
AWS_SDK_LOAD_CONFIG="true" \
AWS_PROFILE="slashdevops" \
./aws_cloudwatch_exporter metrics get \
    --metricsFiles ~/tmp/queries/m1.yaml \
    --debug
```

## Development / Contributing

WIP

## License

This software is released under the APACHE LICENSE, VERSION 2.0:

* [http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

## Author Information

* [Christian González Di Antonio](https://github.com/christiangda)
