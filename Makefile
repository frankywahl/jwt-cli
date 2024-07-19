BINARY = jwt

VERSION?="tip"
COMMIT?=$(shell git rev-parse HEAD)
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GITHUB_TOKEN?=""

LDFLAGS = -ldflags "-X github.com/frankywahl/jwt-cli/cmd.GitRevision=${COMMIT} -X github.com/frankywahl/jwt-cli/cmd.Version=${VERSION} -X github.com/frankywahl/jwt-cli/cmd.CreatedAt=${DATE}"

all: clean test vet linux darwin windows

install:
	go build ${LDFLAGS} -o ${GOPATH}/bin/jwt

test:
	go test -v --race ./...

vet:
	go vet ./...

fmt:
	test -z $$(gofmt -l .) # This will return non-0 if unsuccessful  run `go fmt ./...` to fix

release:
	docker run --rm --privileged \
		-v $(shell pwd):/go/src/github.com/frankywahl/jwt \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/frankywahl/jwt \
		-e GITHUB_TOKEN=${GITHUB_TOKEN} \
		goreleaser/goreleaser release --clean --snapshot

docker:
	docker build --pull --build-arg COMMIT=${COMMIT} --build-arg VERSION=${VERSION} --build-arg DATE=${DATE} -t frankywahl/jwt .
