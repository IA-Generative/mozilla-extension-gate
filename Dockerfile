FROM golang:1.27.3-alpine AS base

WORKDIR /workspace/backend

RUN apk add --no-cache ca-certificates git nodejs npm
RUN go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum ./
RUN go mod download

FROM base AS dev

WORKDIR /app

COPY . ./
COPY .air.toml ./

EXPOSE 8080

CMD ["/bin/sh", "-c", "air -c .air.toml"]

FROM base AS builder

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /tmp/server ./cmd/server
WORKDIR /tmp
RUN mkdir docs
RUN WRITE_SPEC_ONLY=true ./server

FROM alpine:3.18 AS prod

ARG VERSION="unknown"
ENV APP_VERSION=${VERSION}

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /tmp/server /app/server
COPY --from=builder /workspace/backend/config /app/config

EXPOSE 8080

ENTRYPOINT ["/app/server"]
