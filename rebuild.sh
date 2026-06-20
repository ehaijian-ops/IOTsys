#!/bin/bash
# IoT Server - 离线构建 & 部署脚本
# 在 WSL 终端中运行: bash /mnt/e/IOTsys/rebuild.sh

set -e
PASS="hnA10175"
HOST="root@115.191.21.13"
DEPLOY="/root/iot-deploy"

echo "========================================"
echo " IoT Server 部署 (vendor 离线模式)"
echo "========================================"

# 1. 打包
echo ""
echo "[1/4] 打包项目..."
cd /mnt/e/IOTsys
rm -f server_with_vendor.tar.gz
tar -czf server_with_vendor.tar.gz --exclude=.git --exclude=bin --exclude='*.py' --exclude='*.log' server/
ls -lh server_with_vendor.tar.gz

# 2. 上传
echo ""
echo "[2/4] 上传到服务器..."
sshpass -p "$PASS" scp -o StrictHostKeyChecking=no \
  server_with_vendor.tar.gz ${HOST}:${DEPLOY}/

# 3. 构建
echo ""
echo "[3/4] 远程解压 + 构建镜像..."
sshpass -p "$PASS" ssh -o StrictHostKeyChecking=no ${HOST} bash << 'REMOTE'
set -e
cd /root/iot-deploy
rm -rf server
tar -xzf server_with_vendor.tar.gz
echo "  Vendor size: $(du -sh server/vendor | cut -f1)"

cd server
echo "  Building (offline)..."
docker build -t iot-server:latest . 2>&1
echo "  [OK] Build done"
REMOTE

# 4. 启动
echo ""
echo "[4/4] 启动容器..."
sshpass -p "$PASS" ssh -o StrictHostKeyChecking=no ${HOST} bash << 'REMOTE'
set -e
docker stop iot-server 2>/dev/null || true
docker rm iot-server 2>/dev/null || true
docker network create iot-network 2>/dev/null || true

docker run -d --name iot-server \
  --network iot-network \
  --restart unless-stopped \
  -p 8080:8080 -p 7000:7000 \
  -v /root/iot-deploy/server/config:/app/config \
  -e GIN_MODE=release \
  iot-server:latest

sleep 3
echo ""
echo "=== Container Status ==="
docker ps --filter name=iot-server
echo ""
echo "=== Recent Logs ==="
docker logs iot-server --tail 20 2>&1
echo ""
echo "=== Health Check ==="
curl -s http://localhost:8080/api/health 2>&1 || echo "Health endpoint not available yet"
REMOTE

echo ""
echo "========================================"
echo " 部署完成!"
echo " API: http://115.191.21.13:8080"
echo "========================================"
