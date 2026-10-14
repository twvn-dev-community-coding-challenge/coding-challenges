# API image for docker-compose (logs → Docker → Promtail → Loki).
# Uses vendored modules so `docker build` does not call proxy.golang.org (avoids TLS/CA issues behind corporate proxies).
# Refresh after dependency changes: go mod vendor
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-s -w" -o /sms-otp .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=build /sms-otp /app/sms-otp
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/metrics > /dev/null || exit 1
ENTRYPOINT ["/app/sms-otp"]
