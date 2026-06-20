#!/usr/bin/env python3
"""Deploy both backend (with SSE hub) and frontend (with logs page)"""
import paramiko
import sys
import os
import time

HOST = "123.56.161.254"
USER = "root"
PASSWORD = "hndx@N2000"
DEPLOY_BASE = "/root/iot-deploy"

def safe(s):
    if isinstance(s, bytes):
        s = s.decode('utf-8', errors='replace')
    return s.encode('ascii', errors='replace').decode()

def run_ssh(ssh, cmd, timeout=120):
    stdin, stdout, stderr = ssh.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode('utf-8', errors='replace').strip()
    err = stderr.read().decode('utf-8', errors='replace').strip()
    return out, err

def sftp_upload_dir(ssh, local_dir, remote_dir):
    """Upload a directory recursively via SFTP"""
    sftp = ssh.open_sftp()
    count = 0
    for root, dirs, files in os.walk(local_dir):
        for f in files:
            local_path = os.path.join(root, f).replace("\\", "/")
            rel = os.path.relpath(local_path, local_dir).replace("\\", "/")
            remote_path = "{}/{}".format(remote_dir, rel)
            remote_subdir = os.path.dirname(remote_path)
            # Create remote directories
            parts = remote_subdir.strip("/").split("/")
            for i in range(1, len(parts) + 1):
                try:
                    sftp.mkdir("/" + "/".join(parts[:i]))
                except:
                    pass
            sftp.put(local_path, remote_path)
            count += 1
            print("  Uploaded: {}".format(safe(rel)))
    sftp.close()
    return count

def main():
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(HOST, username=USER, password=PASSWORD, timeout=15)

    # ========== Part 1: Deploy Backend ==========
    print("=" * 50)
    print("PART 1: Deploy Backend (with SSE Hub)")
    print("=" * 50)

    # Clean and upload server code
    out, err = run_ssh(ssh, "rm -rf {}/server && mkdir -p {}/server".format(DEPLOY_BASE, DEPLOY_BASE))
    print("Cleaned server dir")

    web_dir = "e:/IOTsys/server"
    n = sftp_upload_dir(ssh, web_dir, "{}/server".format(DEPLOY_BASE))
    print("Uploaded {} server files".format(n))

    # Build
    print("\n>>> Building iot-server image...")
    out, err = run_ssh(ssh, "cd {}/server && docker build --no-cache -t iot-server:latest . 2>&1".format(DEPLOY_BASE), timeout=300)
    print(safe(out))
    if err and "ERROR" in err:
        print("BUILD ERROR:", safe(err[:500]))
        ssh.close()
        sys.exit(1)

    # Stop old, start new
    out, err = run_ssh(ssh, "docker stop iot-server 2>/dev/null; docker rm iot-server 2>/dev/null")
    cmd = ("docker run -d --name iot-server "
           "--network iot-network "
           "-p 8080:8080 -p 7000:7000 "
           "--restart unless-stopped "
           "iot-server:latest")
    out, err = run_ssh(ssh, cmd)
    print("Server container:", safe(out or "started"))

    time.sleep(5)
    # Ensure config is correct
    out, err = run_ssh(ssh, "docker exec iot-server ls /app/config/config.yaml 2>/dev/null || docker exec iot-server cp /app/config/config.prod.yaml /app/config/config.yaml 2>/dev/null; echo ok")
    out, err = run_ssh(ssh, "docker restart iot-server 2>/dev/null")
    time.sleep(3)

    # Verify backend
    out, err = run_ssh(ssh, "curl -s http://localhost:8080/health")
    print("Health check:", safe(out))

    # ========== Part 2: Deploy Frontend ==========
    print("\n" + "=" * 50)
    print("PART 2: Deploy Frontend (with Logs page)")
    print("=" * 50)

    # Clean and upload web admin code
    admin_remote = "{}/web-admin".format(DEPLOY_BASE)
    out, err = run_ssh(ssh, "rm -rf {} && mkdir -p {}".format(admin_remote, admin_remote))
    print("Cleaned web-admin dir")

    web_admin = "e:/IOTsys/web/admin"
    n = sftp_upload_dir(ssh, web_admin, admin_remote)
    print("Uploaded {} web-admin files".format(n))

    # Build
    print("\n>>> Building iot-admin image...")
    out, err = run_ssh(ssh, "cd {} && docker build -t iot-admin:latest . 2>&1".format(admin_remote), timeout=300)
    print(safe(out))
    if err and "ERROR" in err:
        print("BUILD ERROR:", safe(err[:500]))
        ssh.close()
        sys.exit(1)

    # Restart admin
    out, err = run_ssh(ssh, "docker stop iot-admin 2>/dev/null; docker rm iot-admin 2>/dev/null")
    cmd = ("docker run -d --name iot-admin "
           "--network iot-network "
           "-p 8081:80 "
           "--restart unless-stopped "
           "iot-admin:latest")
    out, err = run_ssh(ssh, cmd)
    print("Admin container:", safe(out or "started"))

    time.sleep(3)

    # ========== Final Verify ==========
    print("\n" + "=" * 50)
    print("FINAL VERIFY")
    print("=" * 50)

    out, err = run_ssh(ssh, "docker ps --filter 'name=iot-' --format '{{.Names}} {{.Status}}'")
    print("Containers:\n{}".format(safe(out)))

    out, err = run_ssh(ssh, "curl -s http://localhost:8080/health")
    print("Backend API: {}".format(safe(out)))

    out, err = run_ssh(ssh, "curl -s -o /dev/null -w '%{http_code}' http://localhost:8081/")
    print("Frontend HTTP: {}".format(safe(out)))

    out, err = run_ssh(ssh, "curl -s http://localhost:8081/ | grep -o '<title>[^<]*'")
    print("Frontend title: {}".format(safe(out)))

    print("\n=== DONE ===")
    print("Admin Panel: http://{}:8081".format(HOST))
    print("New page:    http://{}:8081/#/logs".format(HOST))
    print("API Base:    http://{}:8080/api/v1".format(HOST))
    print("SSE Stream:  http://{}:8080/api/v1/devices/logs/stream".format(HOST))
    ssh.close()

if __name__ == "__main__":
    main()
