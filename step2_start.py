import paramiko, time

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('115.191.21.13', username='root', password='hnA10175', timeout=30)

# Check build log
stdin, stdout, stderr = ssh.exec_command('tail -20 /root/iot-deploy/build.log 2>/dev/null || echo "no log yet"')
log = stdout.read().decode()
print("=== Build Log ===")
print(log)

# Check if build succeeded
if 'error' in log.lower() and 'exporting' not in log.lower():
    print("\n*** BUILD FAILED ***")
    ssh.close()
    exit(1)

# Check image exists
stdin, stdout, stderr = ssh.exec_command(
    'docker images iot-server:latest --format "[OK] {{.Repository}}:{{.Tag}} {{.Size}}"')
img = stdout.read().decode().strip()
if not img:
    print("\n*** Image not found, build may still be running ***")
    ssh.close()
    exit(0)
print(img)

# Stop old, start new
stdin, stdout, stderr = ssh.exec_command(
    'docker stop iot-server 2>/dev/null; '
    'docker rm iot-server 2>/dev/null; '
    'docker network create iot-network 2>/dev/null; '
    'docker run -d --name iot-server --network iot-network '
    '--restart unless-stopped -p 8080:8080 -p 7000:7000 '
    '-v /root/iot-deploy/server/config:/app/config '
    '-e GIN_MODE=release iot-server:latest && '
    'sleep 3 && '
    'echo "=== Status ===" && '
    'docker ps --filter name=iot-server && '
    'echo "=== Logs ===" && '
    'docker logs iot-server --tail 30 2>&1'
)
print(stdout.read().decode())

# Verify
time.sleep(2)
stdin, stdout, stderr = ssh.exec_command('curl -s http://localhost:8080/api/health 2>&1')
print("[Health]", stdout.read().decode().strip()[:200])

ssh.close()
