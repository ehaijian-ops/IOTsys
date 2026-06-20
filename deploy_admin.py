#!/usr/bin/env python3
"""Deploy web admin frontend to server"""
import paramiko
import sys
import time

HOST = "123.56.161.254"
USER = "root"
PASSWORD = "hndx@N2000"
REMOTE_DIR = "/root/iot-deploy/web-admin"

def safe(s):
    """Encode string safely for Windows console"""
    if isinstance(s, bytes):
        s = s.decode('utf-8', errors='replace')
    return s.encode('ascii', errors='replace').decode()

def run_ssh(ssh, cmd, timeout=120):
    stdin, stdout, stderr = ssh.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode('utf-8', errors='replace').strip()
    err = stderr.read().decode('utf-8', errors='replace').strip()
    return out, err

def main():
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(HOST, username=USER, password=PASSWORD, timeout=15)

    print("=== Step 1: Clean old web-admin dir ===")
    out, err = run_ssh(ssh, "rm -rf {0} && mkdir -p {0}".format(REMOTE_DIR))
    print(safe(out or "OK"))

    print("\n=== Step 2: Upload web-admin files ===")
    sftp = ssh.open_sftp()
    import os
    web_dir = "e:/IOTsys/web/admin"
    for root, dirs, files in os.walk(web_dir.replace("\\", "/")):
        for f in files:
            local_path = os.path.join(root, f).replace("\\", "/")
            rel = os.path.relpath(local_path, web_dir.replace("\\", "/")).replace("\\", "/")
            remote_path = "{}/{}".format(REMOTE_DIR, rel)
            remote_subdir = os.path.dirname(remote_path)
            try:
                sftp.stat(remote_subdir)
            except:
                parts = remote_subdir.strip("/").split("/")
                for i in range(1, len(parts) + 1):
                    try:
                        sftp.mkdir("/" + "/".join(parts[:i]))
                    except:
                        pass
            sftp.put(local_path, remote_path)
            print("  Uploaded: {}".format(safe(rel)))
    sftp.close()

    print("\n=== Step 3: Build Docker image ===")
    out, err = run_ssh(ssh, "cd {} && docker build -t iot-admin:latest . 2>&1".format(REMOTE_DIR), timeout=300)
    print(safe(out))
    if err and "error" in err.lower():
        print("BUILD ERROR:", safe(err))
        ssh.close()
        sys.exit(1)

    print("\n=== Step 4: Stop old admin container if exists ===")
    out, err = run_ssh(ssh, "docker stop iot-admin 2>/dev/null; docker rm iot-admin 2>/dev/null")
    print(safe(out or "Cleaned"))

    print("\n=== Step 5: Start admin container ===")
    cmd = ("docker run -d --name iot-admin "
           "--network iot-network "
           "-p 8081:80 "
           "--restart unless-stopped "
           "iot-admin:latest")
    out, err = run_ssh(ssh, cmd)
    print(safe(out or "Container started"))

    print("\n=== Step 6: Verify ===")
    time.sleep(3)
    out, err = run_ssh(ssh, "docker ps --filter name=iot-admin --format '{{.Names}} {{.Status}}'")
    print("Container:", safe(out))
    out, err = run_ssh(ssh, "curl -s -o /dev/null -w '%{http_code}' http://localhost:8081/")
    print("HTTP status: {}".format(safe(out)))
    out2, _ = run_ssh(ssh, "curl -s http://localhost:8081/ | head -5")
    print("Page preview: {}".format(safe(out2)[:200]))

    print("\n=== Done! Admin panel: http://{}:8081 ===".format(HOST))
    ssh.close()

if __name__ == "__main__":
    main()
