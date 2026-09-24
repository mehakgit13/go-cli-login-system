FROM golang:1.22-bookworm AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/go-cli-login ./cmd/cli

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /out/go-cli-login /app/go-cli-login
RUN mkdir -p /app/data && chmod 700 /app/data
ENV DB_PATH=/app/data/app.db
ENV SESSION_TIMEOUT=30m
ENV MAX_FAILED_ATTEMPTS=5
ENV LOCKOUT_DURATION=15m
ENV PASSWORD_MIN_LENGTH=8
ENV SESSION_SECRET=change-me-in-production
ENTRYPOINT ["/app/go-cli-login"]
