# Memcached Exporter for Prometheus

A memcached exporter for prometheus.

# Building and Running

The memcache exporter exports metrics from a single memcached server for
consumption by prometheus. The server is specified as an argument to the
program.

By default the memcache\_exporter serves on port `9016` at `/metrics`

```
make
./memcache_exporter memcached:11211
```

Alternatively a Dockerfile is supplied

```
docker build -t memcache_exporter .
docker run memcache_exporter
```

To change the server scraped using the Dockerfile method, simply create your
own Dockerfile, and overwrite the `CMD` setting. This is also the way to enable
logging etc.

```
FROM snapbug/memcached-exporter
CMD ["yourserver:yourport"]
```

# Collectors

The exporter collects a number of collections from the server:

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
