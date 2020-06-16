# aws_cloudwatch_exporter

[![Master branch workflow](https://github.com/slashdevops/aws_cloudwatch_exporter/workflows/Master%20branch%20workflow/badge.svg?branch=master)][github]
[![Go Report Card](https://goreportcard.com/badge/github.com/slashdevops/aws_cloudwatch_exporter)](https://goreportcard.com/report/github.com/slashdevops/aws_cloudwatch_exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/slashdevops/aws_cloudwatch_exporter.svg?maxAge=604800)][hub]

Prometheus exporter for AWS CloudWatch

This exporter used the AWS CloudWatch GetMetricsData API call, please it is important you read
* https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/

## Configuration

The configuration could be set using 3 ways:
 
1. Configuration Files   (i.e.: ./server.yaml, ./credentials.yaml, ./metrics.yaml)
2. Environment Variables (i.e.: SERVER_PORT, AWS_REGION, AWS_PROFILE, etc)
3. Program Flags         (i.e.: --serverPort, --region, --profile)

The precedence is in the same order of the list, so, if you define values into `server.yaml` and then 
the same configuration key is defined as a `Env Var` this last will replace the file value.

### Configuration Files

* [server.yaml](docs/server.md)
* [credentials.yaml](docs/credentials.md)
* [metrics.yaml](docs/metrics.md)


## Development / Contributing

### Building and release tools

* https://github.com/prometheus/promu

## License

This software is released under the APACHE LICENSE, VERSION 2.0:

* [http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

## Author Information

* [Christian González Di Antonio](https://github.com/christiangda)