FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /core ./cmd/core

# Migration runner: goose CLI + the migration SQL files. Used by the
# `migrate` compose service to bring the schema up before the app starts.
FROM golang:1.26-alpine AS migrate
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1
COPY db/core/migrations /migrations
ENTRYPOINT ["/bin/sh", "-c", "goose -dir /migrations postgres \"$CORE_DB_CONNSTRING\" up"]

FROM gcr.io/distroless/static-debian12@sha256:9c346e4be81b5ca7ff31a0d89eaeade58b0f95cfd3baed1f36083ddb47ca3160
USER nonroot:nonroot
COPY --from=builder --chown=nonroot:nonroot /core /core
EXPOSE 8080
ENTRYPOINT ["/core"]
