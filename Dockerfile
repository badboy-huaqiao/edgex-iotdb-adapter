ARG BASE=golang:1.18-alpine3.16
FROM ${BASE} AS builder

ARG ALPINE_PKG_BASE="make git openssh-client gcc libc-dev zeromq-dev libsodium-dev"
ARG ALPINE_PKG_EXTRA=""

ENV GOPROXY https://goproxy.cn

WORKDIR /edgex-iotdb-adapter

# Replicate the APK repository override.
# If it is no longer necessary to avoid the CDN mirros we should consider dropping this as it is brittle.
RUN sed -e 's/dl-cdn[.]alpinelinux.org/dl-4.alpinelinux.org/g' -i~ /etc/apk/repositories
# Install our build time packages.
RUN apk add --update --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}


COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .

ARG MAKE='make build'
RUN $MAKE

FROM alpine:3.16

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2022-2023: vmware'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/dl-4.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache zeromq dumb-init

WORKDIR /
COPY --from=builder /edgex-iotdb-adapter/cmd /
# COPY --from=builder /edgex-iotdb-adapter/LICENSE /
# COPY --from=builder /edgex-iotdb-adapter/Attribution.txt /

EXPOSE 59990

ENTRYPOINT ["/adapter-server"]
# CMD ["--cp=consul://edgex-core-consul:8500", "--confdir=/res", "--registry"]