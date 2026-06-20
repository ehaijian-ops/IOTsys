#!/usr/bin/env python3
"""Start docker build on remote server."""
import paramiko, time, threading

HOST = "115.191.21.13"
USER = "root"
PASS = "hnA10175"

c = paramiko.SSHClient()
c.set_missing_host_key_policy(paramiko.AutoAddPolicy())
c.connect(hostname=HOST, port=22, username=USER, password=PASS, timeout=15, allow_agent=False, look_for_keys=False)

transport = c.get_transport()
channel = transport.open_session()
channel.setblocking(0)
channel.exec_command("cd /root/iot-deploy/server && docker build -t iot-server:latest . 2>&1")

def read_channel(prefix):
    buffer = ""
    while not channel.exit_status_ready():
        if channel.recv_ready():
            try:
                data = channel.recv(4096).decode("utf-8", errors="replace")
                if data:
                    buffer += data
                    lines = buffer.split("\n")
                    for line in lines[:-1]:
                        print(line)
                    buffer = lines[-1]
            except:
                pass
        time.sleep(1)
    # Read remaining
    while channel.recv_ready():
        try:
            data = channel.recv(4096).decode("utf-8", errors="replace")
            buffer += data
        except:
            pass
    if buffer:
        print(buffer)

import sys
sys.stdout.flush()
read_channel("build")
exit_code = channel.recv_exit_status()
print(f"\n=== Build exit code: {exit_code} ===")

if exit_code == 0:
    # Check image
    si, so, se = c.exec_command("docker images iot-server --format '{{.Tag}} {{.Size}}'", timeout=10)
    print("Image:", so.read().decode().strip())

    # Stop old, start new
    c.exec_command("docker stop iot-server 2>/dev/null; docker rm iot-server 2>/dev/null", timeout=10)
    time.sleep(2)
    c.exec_command("docker network create iot-network 2>/dev/null", timeout=10)
    si, so, se = c.exec_command(
        "docker run -d --name iot-server --network iot-network --restart unless-stopped "
        "-p 8080:8080 -p 7000:7000 "
        "-v /root/iot-deploy/server/config:/app/config "
        "-e GIN_MODE=release iot-server:latest", timeout=10)
    print("Container:", so.read().decode().strip())
    time.sleep(3)
    si, so, se = c.exec_command("docker ps --filter name=iot", timeout=10)
    print("Status:", so.read().decode())
    si, so, se = c.exec_command("docker logs iot-server --tail 20 2>&1", timeout=10)
    print("Logs:", so.read().decode())

c.close()
