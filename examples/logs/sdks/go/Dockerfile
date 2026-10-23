FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /app/logs-sdk-go .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=build /app/logs-sdk-go /usr/local/bin/logs-sdk-go

ENTRYPOINT ["/usr/local/bin/logs-sdk-go"]
