FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tasmota-plug-exporter .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /app/tasmota-plug-exporter /tasmota-plug-exporter

EXPOSE 9184

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -q --spider http://localhost:9184/health || exit 1

ENTRYPOINT ["/tasmota-plug-exporter"]
