FROM rust:1.87-bookworm AS builder

WORKDIR /app

COPY Cargo.toml Cargo.lock ./
COPY src ./src

RUN cargo build --release --locked

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install --no-install-recommends -y ca-certificates \
    && groupadd --system app \
    && useradd --system --gid app --home-dir /app app \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder --chown=app:app /app/target/release/helloworld-rust /app/helloworld-rust

ENV SERVICE_NAME=helloworld \
    SERVER_ADDRESS=0.0.0.0 \
    TOKEN=todo \
    ENABLE_TRACES=false \
    ENABLE_METRICS=false \
    ENABLE_LOGS=false

EXPOSE 8080

USER app

CMD ["/app/helloworld-rust"]
