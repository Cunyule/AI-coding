# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS builder

WORKDIR /src

ENV CGO_ENABLED=0

COPY go.mod ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN GOOS="${TARGETOS:-linux}" GOARCH="${TARGETARCH:-amd64}" \
    go build -trimpath -ldflags="-s -w" -o /out/fingerprint-server ./cmd/server && \
    GOOS="${TARGETOS:-linux}" GOARCH="${TARGETARCH:-amd64}" \
    go build -trimpath -ldflags="-s -w" -o /out/fingerprint-client ./cmd/client

FROM alpine:3.22 AS server

RUN apk add --no-cache ca-certificates && \
    addgroup -S fingerprint && \
    adduser -S -G fingerprint fingerprint

WORKDIR /app

COPY --from=builder /out/fingerprint-server /usr/local/bin/fingerprint-server
COPY rules/fingerprints.yaml /app/rules/fingerprints.yaml

ENV SERVER_ADDR=:8080 \
    RULES_FILE=/app/rules/fingerprints.yaml

EXPOSE 8080

USER fingerprint

ENTRYPOINT ["/usr/local/bin/fingerprint-server"]

FROM alpine:3.22 AS client

RUN apk add --no-cache ca-certificates && \
    addgroup -S fingerprint && \
    adduser -S -G fingerprint fingerprint

WORKDIR /app

COPY --from=builder /out/fingerprint-client /usr/local/bin/fingerprint-client

USER fingerprint

ENTRYPOINT ["/usr/local/bin/fingerprint-client"]
