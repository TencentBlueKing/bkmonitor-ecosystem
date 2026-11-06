ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine AS base

ENV LC_ALL=C.UTF-8 \
    LANG=C.UTF-8

FROM base AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bin/app .


FROM alpine:3.20 AS runtime

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /bin/app /app/reporter

ENV API_URL="fixme" \
    TOKEN="fixme"

ENTRYPOINT ["/app/reporter"]
CMD ["-exporter", "http"]
