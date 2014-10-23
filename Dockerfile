FROM golang:1.3.1-onbuild

MAINTAINER Sean Pont <seanpont@gmail.com>

EXPOSE 5000
CMD ["ftpServer", "5000"]
ENTRYPOINT ["go-wrapper", "run"]
