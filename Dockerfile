FROM gliderlabs/alpine:3.2

RUN apk-install git

RUN apk-install go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

COPY bin/proxy-link /usr/bin/proxy-link

WORKDIR /go/src/github.com/convox/proxy
COPY . /go/src/github.com/convox/proxy
RUN go install ./...

ENTRYPOINT ["proxy-link"]
