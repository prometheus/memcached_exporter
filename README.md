# Memcached Exporter for Prometheus

[![Build Status](https://github.com/prometheus/memcached_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/prometheus/memcached_exporter/actions/workflows/ci.yml)
[![Docker Repository on Quay](https://quay.io/repository/prometheus/memcached-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/prom/memcached-exporter.svg?maxAge=604800)][hub]

A [memcached](https://memcached.org/) exporter for Prometheus.

## Building

The memcached exporter exports metrics from a memcached server for
consumption by Prometheus. The server is specified as `--memcached.address` flag
to the program (default is `localhost:11211`).

By default the memcache_exporter serves on port `0.0.0.0:9150` at `/metrics`:

```sh
make
./memcached_exporter
```

Alternatively a Dockerfile is supplied:

```sh
docker run -p 9150:9150 quay.io/prometheus/memcached-exporter:latest
```

## Collectors

The exporter collects a number of statistics from the server.

For supported metrics see the [metrics documentation](metrics.md).

## TLS and basic authentication

The Memcached Exporter supports TLS and basic authentication.

To use TLS and/or basic authentication, you need to pass a configuration file
using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

To use TLS for connections to memcached, use the `--memcached.tls.*` flags.
See `memcached_exporter --help` for details.

## Multi-target

The exporter also supports the [multi-target](https://prometheus.io/docs/guides/multi-target-exporter/) pattern on the `/scrape` endpoint. Example:
```
curl `localhost:9150/scrape?target=memcached-host.company.com:11211
```

An example configuration using [prometheus-elasticache-sd](https://github.com/maxbrunet/prometheus-elasticache-sd):

```yaml
scrape_configs:
  - job_name: "memcached_exporter_targets"
    file_sd_configs:
    - files:
        - /path/to/elasticache.json  # File created by service discovery
    metrics_path: /scrape
    relabel_configs:
      # Filter for memcached cache nodes
      - source_labels: [__meta_elasticache_engine]
        regex: memcached
        action: keep
      # Build Memcached URL to use as target parameter for the exporter
      - source_labels:
          - __meta_elasticache_endpoint_address
          - __meta_elasticache_endpoint_port
        replacement: $1
        separator: ':'
        target_label: __param_target
      # Use Memcached URL as instance label
      - source_labels: [__param_target]
        target_label: instance
      # Set exporter address
      - target_label: __address__
        replacement: memcached-exporter-service.company.com:9151
```

If you are running solely for `multi-target` start the exporter with `--memcached.address=""` to avoid attempting to connect to a non existing memcached host, example:

```
./memcached-exporter --memcached.address=""
```

[buildstatus]: https://circleci.com/gh/prometheus/memcached_exporter/tree/master.svg?style=shield
[circleci]: https://circleci.com/gh/prometheus/memcached_exporter
[hub]: https://hub.docker.com/r/prom/memcached-exporter/
[quay]: https://quay.io/repository/prometheus/memcached-exporter
