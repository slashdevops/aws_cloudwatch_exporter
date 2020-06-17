ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc

LABEL maintainer="Christian González Di Antonio <christian@slashdevops.com>" \
      org.opencontainers.image.authors="Christian González Di Antonio <christian@slashdevops.com>" \
      org.opencontainers.image.url="https://github.com/slashdevops/aws_cloudwatch_exporter" \
      org.opencontainers.image.documentation="https://github.com/slashdevops/aws_cloudwatch_exporter" \
      org.opencontainers.image.source="https://github.com/slashdevops/aws_cloudwatch_exporter" \
      org.opencontainers.image.licenses="Apache License Version 2.0" \
      org.opencontainers.image.title="AWS CloudWatch exporter for prometheus.io" \
      org.opencontainers.image.description="AWS CloudWatch exporter for prometheus.io"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/aws_cloudwatch_exporter  /bin/aws_cloudwatch_exporter

EXPOSE      9690
USER        nobody

HEALTHCHECK CMD wget --spider -S "http://localhost:9690/health" -T 60 2>&1 || exit 1

VOLUME "/etc/aws_cloudwatch_exporter"

ENTRYPOINT  [ "/bin/aws_cloudwatch_exporter" ]