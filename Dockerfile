FROM golang:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o netbird-log-forwarder cmd/netbird-log-forwarder/main.go

FROM alpine:3.19
RUN apk update && apk upgrade && apk add --no-cache bash

WORKDIR /app
# COPY config-docker-compose.json ./
COPY --from=builder /app/netbird-log-forwarder .

# CMD ["bash"]
CMD ["/app/netbird-log-forwarder"]
