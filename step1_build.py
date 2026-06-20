import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('115.191.21.13', username='root', password='hnA10175', timeout=30)

# Clean & extract
stdin, stdout, stderr = ssh.exec_command(
    'cd /root/iot-deploy && rm -rf server && '
    'tar -xzf server_with_vendor.tar.gz && '
    'du -sh server/vendor && '
    'docker ps --format "{{.Names}} {{.Status}}"'
)
print(stdout.read().decode())

# Background build
stdin, stdout, stderr = ssh.exec_command(
    'cd /root/iot-deploy/server && '
    'nohup docker build -t iot-server:latest .'
    ' >/root/iot-deploy/build.log 2>&1 &'
)
print('Build PID:', stdout.read().decode() or 'started')
ssh.close()
print('Build started in background. Use step2 to check.')
