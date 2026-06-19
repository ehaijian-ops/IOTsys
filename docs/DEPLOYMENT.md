# 部署运维手册

> IOTsys 物联网充电桩管理平台 · 部署运维指南 v1.0

---

## 一、快速部署（Docker Compose）

### 1.1 环境要求

| 组件   | 最低要求            |
| ------ | ------------------- |
| OS     | Linux (Ubuntu 20.04+) / macOS |
| CPU    | 4 核                |
| 内存   | 8 GB                |
| 磁盘   | 100 GB SSD          |
| Docker | 24.0+               |
| Docker Compose | 2.20+       |

### 1.2 启动步骤

```bash
# 1. 克隆项目
cd /opt
git clone <repo-url> iot-platform
cd iot-platform

# 2. 复制配置模板
cp server/config/config.example.yaml server/config/config.yaml

# 3. 修改配置文件（至少修改 JWT Secret）
vim server/config/config.yaml

# 4. 启动所有服务
docker compose -f deploy/docker-compose/docker-compose.yml up -d

# 5. 查看服务状态
docker compose -f deploy/docker-compose/docker-compose.yml ps

# 6. 查看日志
docker compose -f deploy/docker-compose/docker-compose.yml logs -f iot-server
```

### 1.3 验证部署

```bash
# 健康检查
curl http://localhost:8080/health

# 就绪检查（含在线连接数）
curl http://localhost:8080/ready

# 预期返回
# {"status": "ok", "tcp_connections": 0}
```

### 1.4 服务端口

| 服务          | 端口  | 说明             |
| ------------- | ----- | ---------------- |
| iot-server    | 8080  | HTTP API         |
| iot-server    | 7000  | TCP 设备接入      |
| MySQL         | 3306  | 关系数据库        |
| MongoDB       | 27017 | 文档数据库        |
| Redis         | 6379  | 缓存             |
| Kafka         | 9092  | 消息队列          |
| Zookeeper     | 2181  | Kafka 依赖        |
| 前端 (开发)    | 3000  | Vue Dev Server   |

---

## 二、生产环境部署（Kubernetes）

### 2.1 架构概览

```
                    ┌──────────────────┐
                    │   Ingress / LB   │
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
        ┌─────┴─────┐  ┌────┴─────┐  ┌─────┴─────┐
        │  Frontend │  │ API (x3) │  │ WebSocket │
        │   Nginx   │  │ Gin Pod  │  │   (x2)    │
        └───────────┘  └────┬─────┘  └───────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
  ┌─────┴─────┐        ┌────┴─────┐        ┌─────┴─────┐
  │   MySQL   │        │  Redis   │        │  MongoDB  │
  │  (主从)   │        │ (Cluster)│        │ (Replica) │
  └───────────┘        └──────────┘        └───────────┘
```

### 2.2 Namespace 与资源规划

```yaml
# namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: iot-platform
---
apiVersion: v1
kind: Namespace
metadata:
  name: iot-infra
```

### 2.3 ConfigMap & Secret

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: iot-server-config
  namespace: iot-platform
data:
  config.yaml: |
    server:
      name: iot-platform
      env: production
      port: 8080
    tcp:
      enabled: true
      port: 7000
      max_connections: 50000
      read_timeout: 60s
      write_timeout: 30s
      heartbeat_interval: 30s
    log:
      level: info
      format: json
      output: stdout
---
apiVersion: v1
kind: Secret
metadata:
  name: iot-server-secret
  namespace: iot-platform
type: Opaque
stringData:
  mysql-password: "<your-mysql-password>"
  redis-password: "<your-redis-password>"
  jwt-secret: "<your-jwt-secret-256bit>"
  mongo-uri: "mongodb://user:pass@mongodb:27017"
```

### 2.4 Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iot-server
  namespace: iot-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app: iot-server
  template:
    metadata:
      labels:
        app: iot-server
    spec:
      containers:
        - name: iot-server
          image: iot-server:latest
          ports:
            - containerPort: 8080
              name: http
            - containerPort: 7000
              name: tcp
          env:
            - name: CONFIG_PATH
              value: /app/config/config.yaml
          envFrom:
            - secretRef:
                name: iot-server-secret
          volumeMounts:
            - name: config
              mountPath: /app/config
          resources:
            requests:
              cpu: "500m"
              memory: "512Mi"
            limits:
              cpu: "2000m"
              memory: "2Gi"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
      volumes:
        - name: config
          configMap:
            name: iot-server-config
```

### 2.5 Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: iot-server
  namespace: iot-platform
spec:
  selector:
    app: iot-server
  ports:
    - name: http
      port: 8080
      targetPort: 8080
    - name: tcp
      port: 7000
      targetPort: 7000
  type: ClusterIP
---
# TCP 设备接入需要 LoadBalancer 或 NodePort
apiVersion: v1
kind: Service
metadata:
  name: iot-server-tcp
  namespace: iot-platform
spec:
  selector:
    app: iot-server
  ports:
    - name: tcp
      port: 7000
      targetPort: 7000
      nodePort: 30700
  type: NodePort
```

### 2.6 HPA（水平自动扩缩容）

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: iot-server-hpa
  namespace: iot-platform
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: iot-server
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

---

## 三、监控方案

### 3.1 指标采集（Prometheus + Grafana）

#### 应用指标（需在代码中暴露 `/metrics` 端点）

| 指标名                          | 类型    | 说明                 |
| ------------------------------- | ------- | -------------------- |
| `iot_tcp_connections_total`     | Gauge   | 当前 TCP 连接数      |
| `iot_tcp_connections_rate`      | Counter | TCP 连接速率          |
| `iot_device_online_total`       | Gauge   | 在线设备数            |
| `iot_device_data_total`         | Counter | 数据上报总数          |
| `iot_command_sent_total`        | Counter | 指令下发总数          |
| `iot_command_latency_seconds`   | Histogram | 指令响应延迟       |
| `iot_decode_errors_total`       | Counter | 协议解码错误数        |
| `iot_kafka_produce_latency`     | Histogram | Kafka 生产延迟     |
| `iot_http_requests_total`       | Counter | HTTP 请求总数         |
| `iot_http_request_duration_seconds` | Histogram | HTTP 请求延迟    |

#### Prometheus 配置

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'iot-server'
    scrape_interval: 15s
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - iot-platform
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: iot-server
```

#### Grafana Dashboard 建议

1. **设备概览**：在线数趋势、连接数、协议分布饼图
2. **数据吞吐**：上报速率、解码错误率、Kafka 积压
3. **指令监控**：下发成功率、P50/P95/P99 延迟
4. **系统资源**：CPU/Memory/GC、Goroutine 数

### 3.2 日志收集（EFK / Loki）

```
应用日志 (stdout JSON)
    │
    ▼
Fluentd / Promtail (DaemonSet)
    │
    ▼
Elasticsearch / Loki
    │
    ▼
Kibana / Grafana
```

#### 日志格式规范

```json
{
  "level": "info",
  "ts": "2024-01-01T12:00:00.000Z",
  "caller": "tcpserver/tcpserver.go:67",
  "msg": "TCP server started",
  "addr": "0.0.0.0:7000",
  "max_connections": 50000
}
```

---

## 四、备份策略

### 4.1 MySQL 备份

```bash
#!/bin/bash
# backup-mysql.sh
BACKUP_DIR="/backup/mysql"
DB_NAME="iot_platform"
DATE=$(date +%Y%m%d_%H%M%S)

# 全量备份（每日凌晨 2:00）
mysqldump -h localhost -u root -p"$MYSQL_PASSWORD" \
  --single-transaction \
  --routines \
  --triggers \
  "$DB_NAME" | gzip > "$BACKUP_DIR/${DB_NAME}_${DATE}.sql.gz"

# 保留最近 30 天的备份
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +30 -delete
```

### 4.2 Redis 备份

```bash
#!/bin/bash
# Redis 通过 RDB + AOF 持久化
# redis.conf:
# save 900 1
# save 300 10
# save 60 10000
# appendonly yes
# appendfsync everysec
```

### 4.3 MongoDB 备份

```bash
#!/bin/bash
# backup-mongo.sh
BACKUP_DIR="/backup/mongodb"
DATE=$(date +%Y%m%d_%H%M%S)

mongodump --uri="$MONGO_URI" \
  --gzip \
  --out "$BACKUP_DIR/mongodump_$DATE"

# 保留最近 7 天的备份
find "$BACKUP_DIR" -name "mongodump_*" -mtime +7 -delete
```

### 4.4 备份验证

```bash
# 定期验证备份可恢复性
# 在测试环境中恢复最新备份
mysql -u root -p"$MYSQL_PASSWORD" iot_platform_test < latest_backup.sql
```

---

## 五、故障处理

### 5.1 常见问题

| 问题                         | 可能原因                        | 处理方式                           |
| ---------------------------- | ------------------------------- | ---------------------------------- |
| 设备无法连接 TCP 7000 端口    | 防火墙未放行 / iptables 规则     | 检查安全组和防火墙规则              |
| 设备频繁上下线               | 心跳超时配置过短 / 网络不稳定    | 增大 `heartbeat_interval`          |
| Kafka 消息积压               | 消费者处理慢 / 分区不足          | 增加消费者实例 / 增加分区数         |
| Redis 内存使用过高           | 实时数据 TTL 未生效 / 设备过多    | 检查 TTL 配置 / 扩容 Redis 内存     |
| MongoDB 写入慢               | 磁盘 IO 瓶颈 / 索引过多          | 使用 SSD / 优化索引                 |
| 指令下发超时                 | 设备响应慢 / 网络延迟            | 增大 `timeout` 参数 / 检查网络      |
| MySQL 连接池耗尽             | 并发请求过多                     | 增大 `max_open_conns` / 检查慢查询  |

### 5.2 诊断命令

```bash
# 检查 TCP 连接数
netstat -an | grep :7000 | wc -l

# 检查 Redis 在线设备
redis-cli SCARD devices:online

# 检查 Kafka 消费延迟
kafka-consumer-groups --bootstrap-server localhost:9092 \
  --group data-processor --describe

# 检查 MongoDB 连接
mongosh --eval "db.serverStatus().connections"

# 检查 Go 程序 goroutine
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

### 5.3 紧急回滚

```bash
# 1. 切换到上一个稳定版本的镜像
kubectl set image deployment/iot-server \
  iot-server=iot-server:v1.2.3 -n iot-platform

# 2. 查看回滚状态
kubectl rollout status deployment/iot-server -n iot-platform

# 3. 如需完全回滚到上一版本
kubectl rollout undo deployment/iot-server -n iot-platform
```

---

## 六、安全加固

### 6.1 网络安全

- 数据库（MySQL/Redis/MongoDB）不暴露公网端口
- TCP 7000 端口仅对设备 IP 段开放
- API 8080 端口通过 Nginx/Ingress 反向代理，启用 HTTPS
- Kafka 启用 SASL/SSL 认证

### 6.2 应用安全

- JWT Token 有效期：Access 2h / Refresh 7d
- API 限流：每 IP 100 req/s
- 敏感配置使用 Kubernetes Secret 或 Vault
- 定期轮换密钥和证书

### 6.3 审计

- 所有管理操作记录到 `audit_logs` 集合
- 包含：操作人、操作类型、目标资源、时间、IP

---

## 七、升级策略

### 7.1 滚动升级

```bash
# 构建新镜像
docker build -t iot-server:v1.1.0 -f server/Dockerfile .

# 更新 Deployment（Kubernetes 自动滚动更新）
kubectl set image deployment/iot-server \
  iot-server=iot-server:v1.1.0 -n iot-platform

# 监控升级过程
kubectl rollout status deployment/iot-server -n iot-platform
```

### 7.2 灰度发布（金丝雀）

```yaml
# 创建 Canary Deployment（10% 流量）
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iot-server-canary
  namespace: iot-platform
spec:
  replicas: 1
  selector:
    matchLabels:
      app: iot-server
      version: canary
  template:
    metadata:
      labels:
        app: iot-server
        version: canary
    spec:
      containers:
        - name: iot-server
          image: iot-server:v1.1.0-rc1
```

---

## 八、容量规划

### 8.1 资源参考

| 设备规模   | API 副本 | CPU/副本 | 内存/副本 | Redis 内存 | MongoDB 磁盘/天 |
| ---------- | -------- | -------- | --------- | ---------- | --------------- |
| < 1,000    | 2        | 500m     | 512Mi     | 1 GB       | 3 GB            |
| 1,000-10,000 | 3      | 1        | 1Gi       | 4 GB       | 30 GB           |
| 10,000-50,000 | 5     | 2        | 2Gi       | 8 GB       | 144 GB          |
| > 50,000   | 10+      | 4        | 4Gi       | 16 GB+     | 需分层存储       |

### 8.2 扩缩容指南

- **水平扩展**：API 服务无状态，直接增加副本
- **TCP 服务**：需要 LoadBalancer 做连接分发，或按设备 ID 哈希分片
- **Redis**：使用 Cluster 模式，按 slot 分片
- **Kafka**：增加分区数（需预先规划，分区只增不减）
- **MongoDB**：使用分片集群（Sharded Cluster）
