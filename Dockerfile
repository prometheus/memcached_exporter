ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/memcached_exporter /bin/memcached_exporter

USER       nobody
ENTRYPOINT ["/bin/memcached_exporter"]
EXPOSE     9150
