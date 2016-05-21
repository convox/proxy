all: build

build:
	docker build --no-cache -t convox/proxy .

test:
	go test -cover -v ./...

release: build
	docker push convox/proxy
