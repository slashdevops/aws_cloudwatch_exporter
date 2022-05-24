ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest

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

ARG PORT="9690"
EXPOSE  ${PORT}

RUN mkdir -p /home/nobody/.aws && chown -R nobody.nogroup /home/nobody
ENV HOME="/home/nobody"
USER nobody

HEALTHCHECK CMD wget --spider -S "http://localhost:${PORT}/health" -T 60 2>&1 || exit 1

VOLUME ["/home/nobody"]

ENTRYPOINT  [ "/bin/aws_cloudwatch_exporter" ]
# CMD  [ "/bin/aws_cloudwatch_exporter" ]