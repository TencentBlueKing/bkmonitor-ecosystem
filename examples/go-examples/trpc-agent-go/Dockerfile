ARG GO_VERSION=1.24
ARG ALPINE_VERSION=3.21

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -o /bin/server .

FROM alpine:${ALPINE_VERSION} AS app

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=build /bin/server /bin/server
COPY trpc_go.yaml ./

USER app

EXPOSE 8080

ENTRYPOINT ["/bin/server"]
