# Memcached Exporter for Prometheus

A memcached exporter for prometheus.

# Building and Running

The memcache exporter uses an environment variable (`memcache_servers`) to
determine which servers to connect to. This needs to be a single string of
comma-separated host:port pairs. Example `host1:port1,host2:port2` will
monitor two memcache servers at `host1:port` and `host2:port2`.

```
make
memcache_servers=memcache:11211 ./memcache_exporter
```

Alternatively a Dockerfile is supplied

```
docker build -t memcache_exporter .
docker run -e memcache_servers=memcache:11211 memcache_exporter
```

By default the memcache\_exporter serves on port `9016` at `/metrics`

# Collectors

The exporter collects a number of collections for each server:

- `up`: whether the server is up.

- `uptime`: how long the server has been up.

- `cache`: exposes the number of cache hits and misses for
	each server and command. For instance `{command='get',status='hits'}`
	will say how many `get` commands resulted in a hit in the cache.

- `bytes`: exposes the number of bytes read and written by each
	server, under the label `direction`.

- `removal`: exposes how many keys have been expired and evicted.
	In the case of evicted keys it's also separated by whether they were
	ever fetched or not.

- `usage`: exposes the current and total number of connections to the cache
	and items in the cache.
