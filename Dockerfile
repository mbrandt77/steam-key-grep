FROM golang:1.16.3-alpine3.13 as build-env

RUN apk add --no-cache git build-base

ENV GO111MODULE=on
ADD . /tmp/src


WORKDIR /tmp/src

RUN go build -o /tmp/src/build/steamkeygrep cmd/steamkeygrep/main.go && \
    go test -v ./...

FROM alpine:3.13

COPY --from=build-env /tmp/src/build/steamkeygrep /usr/local/bin/steamkeygrep
COPY --from=build-env /tmp/src/conf.yaml conf.yaml

ENTRYPOINT [ "steamkeygrep" ]