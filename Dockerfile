FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /core ./cmd/core

FROM gcr.io/distroless/static-debian12
USER nonroot:nonroot
COPY --from=builder --chown=nonroot:nonroot /core /core
EXPOSE 8080
ENTRYPOINT ["/core"]
