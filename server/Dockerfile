FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy vendor directory and source code (offline build, no network needed)
COPY . .

# Build with vendored dependencies (no network required)
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-w -s" -o /iot-server ./cmd/main.go

# Runtime stage
FROM alpine:3.19

# Use Aliyun mirror for apk
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /iot-server .
COPY config/config.prod.yaml config/config.yaml

EXPOSE 8080 7000

CMD ["./iot-server"]
