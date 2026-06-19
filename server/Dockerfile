FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /iot-server ./cmd/main.go

# 运行阶段
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /iot-server .
COPY config/config.example.yaml /app/config/config.yaml

EXPOSE 8080 7000

CMD ["./iot-server"]
