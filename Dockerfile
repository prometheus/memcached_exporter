FROM golang:1.6-onbuild
MAINTAINER Matt Crane <mcrane@snapbug.geek.nz>

ENTRYPOINT [ "go-wrapper", "run" ]
EXPOSE 9106
