#!/usr/bin/env python3
"""Deploy only the backend (server/) to remote server.
Optimized: tar+gzip upload instead of per-file SFTP (2787 vendor files → single archive).
"""
import paramiko
import sys
import os
import time
import tarfile
import io

HOST = "123.56.161.254"
USER = "root"
PASSWORD = "hndx@N2000"
DEPLOY_BASE = "/root/iot-deploy"

# Files/dirs to exclude from upload (already present in vendor or build artifacts)
EXCLUDE_PATTERNS = {"node_modules", ".git", "__pycache__"}

def safe(s):
    if isinstance(s, bytes):
        s = s.decode('utf-8', errors='replace')
    return s.encode('ascii', errors='replace').decode()

def run_ssh(ssh, cmd, timeout=300):
    stdin, stdout, stderr = ssh.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode('utf-8', errors='replace').strip()
    err = stderr.read().decode('utf-8', errors='replace').strip()
    return out, err

def create_tar_gz(local_dir):
    """Create an in-memory tar.gz archive of local_dir, excluding build artifacts."""
    buf = io.BytesIO()
    with tarfile.open(fileobj=buf, mode='w:gz') as tar:
        for root, dirs, files in os.walk(local_dir):
            # Skip excluded directories
            dirs[:] = [d for d in dirs if d not in EXCLUDE_PATTERNS]
            for f in files:
                local_path = os.path.join(root, f).replace("\\", "/")
                rel = os.path.relpath(local_path, local_dir).replace("\\", "/")
                # Skip excluded files
                skip = False
                for pat in EXCLUDE_PATTERNS:
                    if pat in rel.split("/"):
                        skip = True
                        break
                if skip:
                    continue
                tar.add(local_path, arcname=rel)
    buf.seek(0)
    return buf

def upload_tar(ssh, tar_data, remote_dir):
    """Upload tar.gz data via SFTP, then extract on remote server."""
    sftp = ssh.open_sftp()
    remote_tar = "{}/upload.tar.gz".format(remote_dir)
    sftp.putfo(tar_data, remote_tar)
    size_mb = tar_data.tell() / (1024 * 1024)
    print("  Uploaded archive: {:.1f} MB".format(size_mb))
    sftp.close()

    # Extract on remote
    out, err = run_ssh(ssh, "cd {} && tar xzf upload.tar.gz && rm upload.tar.gz && echo OK".format(remote_dir))
    print("  Extract:", safe(out))
    if err:
        print("  Extract stderr:", safe(err))

def main():
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    ssh.connect(HOST, username=USER, password=PASSWORD, timeout=15)

    print("=" * 50)
    print("Deploy Backend (Optimized tar.gz upload)")
    print("=" * 50)

    # Clean remote server dir
    out, err = run_ssh(ssh, "rm -rf {}/server && mkdir -p {}/server".format(DEPLOY_BASE, DEPLOY_BASE))
    print("Cleaned server dir")

    # Create tar.gz and upload
    local_dir = "e:/IOTsys/server"
    print("Packing {} ...".format(local_dir))
    tar_data = create_tar_gz(local_dir)
    upload_tar(ssh, tar_data, DEPLOY_BASE + "/server")

    # Build Docker image (use cache to speed up, only code changes trigger rebuild)
    print("\n>>> Building iot-server image...")
    out, err = run_ssh(ssh, "cd {}/server && docker build -t iot-server:latest . 2>&1".format(DEPLOY_BASE), timeout=600)
    print(safe(out[-4000:]))  # Show last 4000 chars
    if err and "ERROR" in err:
        print("BUILD STDERR:", safe(err[:1000]))

    # Stop existing container
    print("\n>>> Stopping old iot-server...")
    out, err = run_ssh(ssh, "docker stop iot-server 2>/dev/null; docker rm iot-server 2>/dev/null; echo cleaned")
    print(safe(out))

    # Start new container
    print("\n>>> Starting new iot-server...")
    cmd = ("docker run -d --name iot-server "
           "--network iot-network "
           "-p 8080:8080 -p 7000:7000 "
           "--restart unless-stopped "
           "-e GIN_MODE=release "
           "iot-server:latest")
    out, err = run_ssh(ssh, cmd, timeout=15)
    print("Container ID:", safe(out or "started"))

    time.sleep(5)

    # Verify
    print("\n" + "=" * 50)
    print("VERIFY")
    print("=" * 50)

    out, err = run_ssh(ssh, "docker ps --filter 'name=iot-server' --format '{{.Names}} {{.Status}} {{.CreatedAt}}'")
    print("Container:", safe(out))

    out, err = run_ssh(ssh, "curl -s http://localhost:8080/health")
    print("Health:", safe(out))

    # Check logs for WSD adapter registration
    out, err = run_ssh(ssh, "docker logs iot-server --tail 50 2>&1 | grep -i -E 'WSD|wsd|adapter|protocol'")
    print("\nAdapter Logs:")
    print(safe(out) if out.strip() else "(no adapter logs found)")

    # Full logs tail
    out, err = run_ssh(ssh, "docker logs iot-server --tail 30 2>&1")
    print("\n--- Last 30 lines ---")
    print(safe(out))

    print("\n=== DONE ===")
    print("Watch logs: ssh root@{} 'docker logs iot-server --tail 50 -f'".format(HOST))
    ssh.close()

if __name__ == "__main__":
    main()
