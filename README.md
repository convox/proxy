# convox/proxy

Simple TCP proxy to simulate TLS and PROXY protocol locally

## Usage

    $ docker run -link <container>:host -p <outer-port> convox/proxy <outer-port> <container-port> <protocol> [options...]

#### Docker Host (80, HTTP) to container (5000, HTTP) named `foo`

    $ docker run -link foo:host -p 80 convox/proxy 443 5000 https

#### Docker Host (443, HTTPS) to container (5000, HTTP) named `foo`

    $ docker run -link foo:host -p 443 convox/proxy 443 5000 https

#### Docker Host (443, HTTPS) to container (5001, HTTPS) named `foo`

    $ docker run -link foo:host -p 443 convox/proxy 443 5000 https secure

#### Docker Host (443, TLS) to container (5002, TLS) with PROXY protocol named `foo`

    $ docker run -link foo:host -p 443 convox/proxy 443 5002 tls secure proxy
