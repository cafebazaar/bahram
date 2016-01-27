PLATFORM_ARGS ?= GOOS=linux GOARCH=amd64
VERSION ?= $(shell git describe --tags)
COMMIT := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell LANG=en_US date +"%F_%T_%z")
DOCKER_IMAGE ?= "cafebazaar/bahram"

.PHONY: help clean docker push test
help:
	@echo "Please use \`make <target>' where <target> is one of"
	@echo "  bahram   to build the main binary (for linux/amd64)"
	@echo "  docker   to build the docker image"
	@echo "  push     to push the built docker to docker hub"
	@echo "  test     to run unittests"
	@echo "  clean    to remove generated files"

test: *.go */*.go
	go get -t -v ./...
	go test -v ./...

bahram: *.go */*.go
	go get -v
	$(PLATFORM_ARGS) go build -ldflags "-s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)" -o bahram

clean:
	rm -rf bahram

docker: bahram
	docker build -t $(DOCKER_IMAGE) .

push: docker
	docker push $(DOCKER_IMAGE)
