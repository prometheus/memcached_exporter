FROM        quay.io/prometheus/busybox:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

COPY memcached_exporter /bin/memcached_exporter

USER       nobody
ENTRYPOINT ["/bin/memcached_exporter"]
EXPOSE     9150
