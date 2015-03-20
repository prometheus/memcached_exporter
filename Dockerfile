FROM golang:1.4-onbuild

EXPOSE 9106

ENTRYPOINT ["./memcache_exporter", "-logtostderr"]
