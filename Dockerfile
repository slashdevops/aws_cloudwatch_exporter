ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc
LABEL maintainer="Christian González Di Antonio <christian@slashdevops.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/aws_cloudwatch_exporter  /bin/aws_cloudwatch_exporter

EXPOSE      9690
USER        nobody
ENTRYPOINT  [ "/bin/aws_cloudwatch_exporter" ]