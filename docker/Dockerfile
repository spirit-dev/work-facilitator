FROM --platform=$BUILDPLATFORM golang:alpine@sha256:eb61c3873c97416fff0116553fff65d67a3549f72e11d9758849a689cdd8b523 AS build

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

FROM golangci/golangci-lint:v2.15.0-alpine@sha256:2ba875b990f1d8501ad4ec19be1ea0701e316ec3a4b764f7b129f62f01318bbf AS lint

# Mainly used to debug lint gitlab ci stage

COPY ./ /go/

ENTRYPOINT ["tail", "-f", "/dev/null"]

FROM golang:alpine@sha256:eb61c3873c97416fff0116553fff65d67a3549f72e11d9758849a689cdd8b523 AS build2

# Mainly used to debug mod tidy, mod update, build, build_all gitlab stages

COPY ./ /go/

ENTRYPOINT ["tail", "-f", "/dev/null"]
