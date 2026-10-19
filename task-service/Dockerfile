FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /task-service ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /task-service /task-service
USER nonroot:nonroot
EXPOSE 8081 50052
ENTRYPOINT ["/task-service"]
