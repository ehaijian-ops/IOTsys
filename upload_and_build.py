#!/usr/bin/env python3
"""Upload server code with vendor deps and build IoT Server on remote."""
import paramiko, os, time

HOST = "115.191.21.13"
USER = "root"
PASS = "hnA10175"

c = paramiko.SSHClient()
c.set_missing_host_key_policy(paramiko.AutoAddPolicy())
c.connect(hostname=HOST, port=22, username=USER, password=PASS, timeout=15, allow_agent=False, look_for_keys=False)

# Check server state
print("=== Server State ===")
si, so, se = c.exec_command("docker ps -a --format 'table {{.Names}}\t{{.Status}}' && echo '---' && free -h | head -2", timeout=10)
print(so.read().decode())

# Upload tar
print("=== Uploading server.tar.gz (8.5MB)... ===")
sftp = c.open_sftp()
sftp.put(r"e:\IOTsys\server_with_vendor.tar.gz", "/root/iot-deploy/server_with_vendor.tar.gz")
sftp.close()
print("Upload complete!")

# Extract
print("=== Extracting... ===")
si, so, se = c.exec_command("cd /root/iot-deploy && rm -rf server && tar -xzf server_with_vendor.tar.gz && ls server/vendor/ | head -5 && echo '---' && du -sh server/vendor", timeout=30)
print(so.read().decode())
print(se.read().decode())

# Start MySQL if not running
print("=== Starting DB containers ===")
si, so, se = c.exec_command("docker ps --format '{{.Names}}' | grep -q iot-mysql && echo 'MySQL running' || docker start iot-mysql 2>&1", timeout=30)
print(so.read().decode())
si, so, se = c.exec_command("docker ps --format '{{.Names}}' | grep -q iot-redis && echo 'Redis running' || docker start iot-redis 2>&1", timeout=30)
print(so.read().decode())

# Wait for MySQL to be ready
print("=== Waiting for MySQL... ===")
si, so, se = c.exec_command("for i in $(seq 1 30); do docker exec iot-mysql mysqladmin ping -h localhost -u root -piot_password_2024 2>/dev/null && echo 'MySQL ready' && break; echo 'Waiting...'; sleep 2; done", timeout=90)
print(so.read().decode())

# Build Docker image
print("=== Building IoT Server Docker image (offline, no network)... ===")
si, so, se = c.exec_command(
    "cd /root/iot-deploy/server && docker build -t iot-server:latest . 2>&1",
    timeout=600
)
build_log = so.read().decode()
build_err = se.read().decode()
print(build_log[-2000:])
if build_err:
    print("STDERR:", build_err[-500:])

# Check build result
si, so, se = c.exec_command("docker images iot-server --format '{{.Repository}}:{{.Tag}} {{.Size}} {{.CreatedAt}}'", timeout=10)
img_info = so.read().decode()
print(f"\n=== Image: {img_info.strip()} ===")

if "iot-server" in img_info:
    # Stop old container
    si, so, se = c.exec_command("docker stop iot-server 2>/dev/null; docker rm iot-server 2>/dev/null; echo cleaned", timeout=10)
    print(so.read().decode())

    # Create network if not exists
    si, so, se = c.exec_command("docker network create iot-network 2>/dev/null; echo network_ok", timeout=10)
    print(so.read().decode())

    # Start IoT Server
    print("=== Starting IoT Server ===")
    si, so, se = c.exec_command(
        "docker run -d --name iot-server "
        "--network iot-network "
        "--restart unless-stopped "
        "-p 8080:8080 -p 7000:7000 "
        "-v /root/iot-deploy/server/config:/app/config "
        "-e GIN_MODE=release "
        "iot-server:latest",
        timeout=10
    )
    print(so.read().decode())

    time.sleep(3)
    si, so, se = c.exec_command("docker ps --format '{{.Names}}\t{{.Status}}\t{{.Ports}}' | grep iot", timeout=10)
    print("\n=== Containers: ===")
    print(so.read().decode())

    si, so, se = c.exec_command("docker logs iot-server --tail 20 2>&1", timeout=10)
    print("\n=== IoT Server Logs: ===")
    print(so.read().decode())
else:
    print("BUILD FAILED - check build log above!")

c.close()
