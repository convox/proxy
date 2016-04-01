FROM gliderlabs/alpine:3.2

RUN apk-install git

RUN apk-install go
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

WORKDIR /go/src/github.com/convox/proxy
COPY . /go/src/github.com/convox/proxy
RUN go install ./...

ENTRYPOINT ["/go/bin/proxy"]
