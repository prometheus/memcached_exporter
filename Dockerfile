FROM golang:1.4-onbuild
MAINTAINER Matt Crane <mcrane@snapbug.geek.nz>

ENTRYPOINT [ "go-wrapper", "run" ]
CMD ["-logtostderr"]
EXPOSE 9106
