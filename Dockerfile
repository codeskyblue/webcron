FROM google/golang:latest
MAINTAINER codeskyblue@gmail.com

COPY . /gopath/src/github.com/codeskyblue/webcrontab
WORKDIR /gopath/src/github.com/codeskyblue/webcrontab

RUN go get -v
RUN go build

EXPOSE 80
ENTRYPOINT []
CMD ["./webcrontab", "-port", "80"]

