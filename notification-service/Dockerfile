FROM golang:1.26.11-alpine AS builder

WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /bin/notification-service \
    ./cmd/server


FROM alpine:3.22

RUN apk add --no-cache \
    ca-certificates \
    tzdata

RUN addgroup -S app && \
    adduser -S app -G app

WORKDIR /app

COPY --from=builder \
    /bin/notification-service \
    /usr/local/bin/notification-service

USER app

EXPOSE 8082

ENTRYPOINT ["/usr/local/bin/notification-service"]