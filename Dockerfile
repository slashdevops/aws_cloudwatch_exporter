# Multi-stage docker
FROM quay.io/prometheus/golang-builder AS builder

ADD .   /go/src/github.com/slashdevops/aws_cloudwatch_exporter
WORKDIR go/src/github.com/slashdevops/aws_cloudwatch_exporter

RUN make

# Final Image
FROM        quay.io/prometheus/busybox:glibc
MAINTAINER  The Prometheus Authors <prometheus-developers@googlegroups.com>
COPY        --from=builder go/src/github.com/slashdevops/aws_cloudwatch_exporter  /bin/aws_cloudwatch_exporter

EXPOSE      9559
ENTRYPOINT  [ "/bin/aws_cloudwatch_exporter" ]