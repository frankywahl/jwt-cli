FROM golang:1.15-apline as build-env
ARG VERSION="tip"
ARG COMMIT="HEAD"
ARG DATE=""
RUN apk add --update make
WORKDIR $GOPATH/src/github.com/user/app
ENV VERSION=${VERSION} COMMIT=${COMMIT}
COPY . .
RUN make install

FROM alpine
COPY --from=build-env /go/bin/jwt /usr/local/bin/.
ENTRYPOINT ["jwt"]
