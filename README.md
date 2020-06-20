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

### Running

#### Docker

```bash
make && \
make promu && \
promu build --prefix .build/darwin-amd64 && \
make docker DOCKER_REPO=docker.io/slashdevops

docker run --rm \
    -v ~/tmp/queries/m1.yaml:/metrics/m1.yaml \
    -v ~/.aws:/credentials \
    -v ~/.aws:/credentials \
    -e "AWS_SDK_LOAD_CONFIG=1" \
    -e "AWS_CONFIG_FILE=/credentials/.aws/config" \
    -e "AWS_SHARED_CREDENTIALS_FILE=/credentials/.aws/credentials" \
    -e "AWS_PROFILE=slashdevops" \
    slashdevops/aws-cloudwatch-exporter-linux-amd64:develop metrics get --metricsFiles /metrics/m1.yaml \
    --debug

docker run --rm \
    -v ~/tmp/queries/m1.yaml:/metrics/m1.yaml \
    -v ~/.aws/credentials:/home/.aws/credentials:ro \
    -v ~/.aws/config:/home/.aws/config:ro \
    -e "AWS_PROFILE=slashdevops" \
    slashdevops/aws-cloudwatch-exporter-linux-amd64:develop metrics get --metricsFiles /metrics/m1.yaml \
    --debug

docker run --rm \
    -v ~/tmp/queries/m1.yaml:/metrics/m1.yaml \
    -e "AWS_REGION=eu-west-1" \
    slashdevops/aws-cloudwatch-exporter-linux-amd64:develop metrics get --metricsFiles /metrics/m1.yaml \
    --debug

docker run --rm \
    -v ~/tmp/queries/m1.yaml:/metrics/m1.yaml \
    -v $HOME/.aws/credentials:/home/nobody/.aws/credentials -u nobody \
    -v $HOME/.aws/config:/home/nobody/.aws/config -u nobody \
    slashdevops/aws-cloudwatch-exporter-linux-amd64:develop ls -la /home/nobody/
```

#### Binary

```bash
make

AWS_SDK_LOAD_CONFIG="true" \
AWS_PROFILE="slashdevops" \
./aws_cloudwatch_exporter metrics get --metricsFiles ~/tmp/queries/m1.yaml \
    --debug
```

```bash
./aws_cloudwatch_exporter metrics get --metricsFiles ~/tmp/queries/m1.yaml \
    --profile slashdevops \
    --debug
```

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