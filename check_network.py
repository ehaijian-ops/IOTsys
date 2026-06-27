"""Check Docker network and iptables rules on remote server."""
import paramiko

ssh = paramiko.SSHClient()
ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
ssh.connect('123.56.161.254', username='root', password='hndx@N2000', timeout=15)

# Docker network mode
stdin, stdout, stderr = ssh.exec_command('docker inspect iot-server --format="{{.HostConfig.NetworkMode}}"')
print("Network mode:", stdout.read().decode().strip())

# Container IP
stdin, stdout, stderr = ssh.exec_command('docker inspect iot-server --format="{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}"')
print("Container IP:", stdout.read().decode().strip())

# iptables NAT rules for port 7000
stdin, stdout, stderr = ssh.exec_command('iptables -t nat -L DOCKER -n 2>/dev/null | grep 7000; echo "---done---"')
print("iptables NAT for 7000:", stdout.read().decode().strip())

# Check if there's a current TCP connection on 7000
stdin, stdout, stderr = ssh.exec_command('ss -tnp | grep 7000')
print("TCP connections on 7000:", stdout.read().decode().strip() or "(none)")

# Docker port mappings
stdin, stdout, stderr = ssh.exec_command('docker port iot-server')
print("Port mappings:", stdout.read().decode().strip())

ssh.close()
