FROM golang:1.26.9-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/user-service ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/user-service /user-service
EXPOSE 8080 50051
USER nonroot:nonroot
ENTRYPOINT ["/user-service"]
