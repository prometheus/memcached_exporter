FROM golang:1.4-onbuild
MAINTAINER Matt Crane <mcrane@snapbug.geek.nz>

ENTRYPOINT [ "go-wrapper", "run" ]
CMD ["memcached:11211"]
EXPOSE 9106
