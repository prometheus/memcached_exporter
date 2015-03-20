FROM golang:1.4-onbuild

ENTRYPOINT ["./memcache_exporter", "-logtostderr", "memcache:11211"]
