# Memcached Exporter for Prometheus [![Build Status][buildstatus]][circleci]

[![Docker Repository on Quay](https://quay.io/repository/prometheus/memcached-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/prom/memcached-exporter.svg?maxAge=604800)][hub]

A [memcached](https://memcached.org/) exporter for prometheus.

## Building

The memcached exporter exports metrics from a memcached server for
consumption by prometheus. The server is specified as `-memcached.address` flag
to the program (default is `localhost:11211`).

By default the memcache\_exporter serves on port `0.0.0.0:9106` at `/metrics`

```
make
./memcached_exporter
```

Alternatively a Dockerfile is supplied

```
docker build -t memcached_exporter .
docker run memcached_exporter
```

## Collectors

The exporter collects a number of statistics from the server:

```
# HELP memcached_commands_total Total number of all requests broken down by command (get, set, etc.) and status.
# TYPE memcached_commands_total counter
# HELP memcached_connections_total Total number of connections opened since the server started running.
# TYPE memcached_connections_total counter
# HELP memcached_current_bytes Current number of bytes used to store items.
# TYPE memcached_current_bytes gauge
# HELP memcached_current_connections Current number of open connections.
# TYPE memcached_current_connections gauge
# HELP memcached_current_items Current number of items stored by this instance.
# TYPE memcached_current_items gauge
# HELP memcached_items_evicted_total Number of valid items removed from cache to free memory for new items.
# TYPE memcached_items_evicted_total counter
# HELP memcached_items_reclaimed_total Number of times an entry was stored using memory from an expired entry.
# TYPE memcached_items_reclaimed_total counter
# HELP memcached_items_total Total number of items stored during the life of this instance.
# TYPE memcached_items_total counter
# HELP memcached_limit_bytes Number of bytes this server is allowed to use for storage.
# TYPE memcached_limit_bytes gauge
# HELP memcached_read_bytes_total Total number of bytes read by this server from network.
# TYPE memcached_read_bytes_total counter
# HELP memcached_up Could the memcached server be reached.
# TYPE memcached_up gauge
# HELP memcached_uptime_seconds Number of seconds since the server started.
# TYPE memcached_uptime_seconds counter
# HELP memcached_version The version of this memcached server.
# TYPE memcached_version gauge
# HELP memcached_written_bytes_total Total number of bytes sent by this server to network.
# TYPE memcached_written_bytes_total counter
```

There is also optional support to export metrics about the memcached process
itself by setting the `-memcached.pid-file <path>` flag. If the
memcached\_exporter process has the rights to read /proc information of the
memcached process, then the following metrics will be exported additionally.

```
# HELP memcached_process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE memcached_process_cpu_seconds_total counter
# HELP memcached_process_max_fds Maximum number of open file descriptors.
# TYPE memcached_process_max_fds gauge
# HELP memcached_process_open_fds Number of open file descriptors.
# TYPE memcached_process_open_fds gauge
# HELP memcached_process_resident_memory_bytes Resident memory size in bytes.
# TYPE memcached_process_resident_memory_bytes gauge
# HELP memcached_process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE memcached_process_start_time_seconds gauge
# HELP memcached_process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE memcached_process_virtual_memory_bytes gauge
```

[buildstatus]: https://circleci.com/gh/prometheus/memcached_exporter/tree/master.svg?style=shield
[circleci]: https://circleci.com/gh/prometheus/memcached_exporter
[hub]: https://hub.docker.com/r/prom/node-exporter/
[quay]: https://quay.io/repository/prometheus/node-exporter
