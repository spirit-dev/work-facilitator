FROM --platform=$BUILDPLATFORM golang:alpine AS build

LABEL spiritdev_official="yes" \
    layer="1" \
    from="golang:alpine" \
    maintainer="bordat.jean@gmail.com"

ARG TARGETOS \
    TARGETARCH \
    STL_NAME

WORKDIR /code

COPY ./ /code

RUN addgroup -S application -g 1000 && adduser -S application -u 1000 -G application && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go mod tidy -C src/${STL_NAME} && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -C /code/src/${STL_NAME} -o /code/dist/${STL_NAME}-${TARGETARCH}

ENTRYPOINT ["tail", "-f", "/dev/null"]

FROM golangci/golangci-lint:v1.62.0-alpine@sha256:a94c2da655a1f3e66bb7d2c9232324def349f2ba56a3d715d89e83d98f5bd933 AS lint

# Mainly used to debug lint gitlab ci stage

COPY ./ /go/

ENTRYPOINT ["tail", "-f", "/dev/null"]

FROM golang:alpine AS build2

# Mainly used to debug mod tidy, mod update, build, build_all gitlab stages

COPY ./ /go/

ENTRYPOINT ["tail", "-f", "/dev/null"]
