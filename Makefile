BINARY = jwt
VET_REPORT = vet.report
TEST_REPORT = tests.xml

VERSION="0.0.1"
GOARCH = amd64

COMMIT=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

GITHUB_USERNAME=frankywahl

LDFLAGS = -ldflags "-X github.com/frankywahl/jwt-cli/cmd.GitRevision=${COMMIT} -X github.com/frankywahl/jwt-cli/cmd.Version=${VERSION}"

all: clean test vet linux darwin windows

install:
	go build ${LDFLAGS} -o jwt
	mv jwt ${GOPATH}/bin

darwin:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-darwin-${GOARCH} . ;

linux:
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-linux-${GOARCH} . ;

windows:
	GOOS=windows GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY}-windows-${GOARCH}.exe . ;

test:
	go test -v ./... 2>&1 | tee ${TEST_REPORT} ;

vet:
	go vet ./... > ${VET_REPORT} 2>&1 ;

fmt:
	go fmt $$(go list ./... | grep -v /vendor/) ;



clean:
	-rm -f ${TEST_REPORT}
	-rm -f ${VET_REPORT}
	-rm -f ${BINARY}
	-rm -f ${BINARY}-*

